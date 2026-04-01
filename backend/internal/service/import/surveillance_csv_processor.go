package service

import (
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/google/uuid"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	"github.com/moh-sso-dashboard/internal/model"
	documentRepository "github.com/moh-sso-dashboard/internal/repository/document"
	processRepository "github.com/moh-sso-dashboard/internal/repository/processes"
	"github.com/moh-sso-dashboard/internal/repository/surveillance/interfaces"
	"github.com/moh-sso-dashboard/internal/storage"
)

type SurveillanceCSVProcessor struct {
	documentRepository documentRepository.DocumentRepository
	processRepository  processRepository.ProcessRepository
	importRepository   interfaces.ImportRepository
	storage            storage.Storage
}

func NewSurveillanceCSVProcessor(
	documentRepository documentRepository.DocumentRepository,
	processRepository processRepository.ProcessRepository,
	importRepository interfaces.ImportRepository,
	storage storage.Storage,
) *SurveillanceCSVProcessor {
	return &SurveillanceCSVProcessor{
		documentRepository: documentRepository,
		processRepository:  processRepository,
		importRepository:   importRepository,
		storage:            storage,
	}
}

func (c *SurveillanceCSVProcessor) Process(ctx context.Context, p db.Process) error {
	var (
		batchID      uuid.UUID
		batchCreated bool
	)

	failBatch := func(err error) error {
		if batchCreated {
			notes := err.Error()
			_, _ = c.importRepository.UpdateImportBatchStatus(ctx, db.UpdateImportBatchStatusParams{
				ID:     batchID,
				Status: "FAILED",
				Notes: sql.NullString{
					String: notes,
					Valid:  true,
				},
			})
		}
		return err
	}

	msg := "Loading document"
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 2, &msg)

	doc, err := c.documentRepository.GetDocument(ctx, p.DocumentID)
	if err != nil {
		return err
	}

	msg = "Opening file"
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 5, &msg)

	fileReader, err := c.storage.Download(ctx, doc.ObjectKey)
	if err != nil {
		return err
	}
	defer fileReader.Close()

	msg = "Parsing CSV headers"
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 10, &msg)

	reader := csv.NewReader(fileReader)
	reader.TrimLeadingSpace = true
	reader.ReuseRecord = false

	headers, err := reader.Read()
	if err != nil {
		return err
	}

	for i := range headers {
		headers[i] = normalizeHeader(headers[i])
	}

	if len(headers) == 0 {
		return fmt.Errorf("csv file has no headers")
	}

	datasetType := inferDatasetType(headers)
	if datasetType == "unknown" {
		return fmt.Errorf("unable to infer dataset type from csv headers")
	}

	msg = "Creating import batch"
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 15, &msg)

	batch, err := c.importRepository.CreateImportBatch(ctx, db.CreateImportBatchParams{
		SourceName: "document_upload",
		FileName: sql.NullString{
			String: strings.TrimSpace(doc.OriginalFilename),
			Valid:  strings.TrimSpace(doc.OriginalFilename) != "",
		},
		DatasetType: datasetType,
		ImportedBy: sql.NullString{
			String: p.CreatedBy.String(),
			Valid:  p.CreatedBy != uuid.Nil,
		},
		Status: "PROCESSING",
		Notes: sql.NullString{
			String: "Raw import in progress",
			Valid:  true,
		},
	})
	if err != nil {
		return err
	}

	// Mark as created immediately after successful batch insert
	batchID = batch.ID
	batchCreated = true

	msg = "Importing rows"
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 20, &msg)

	rowCount := 0
	rowNumber := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return failBatch(fmt.Errorf("failed to read csv row %d: %w", rowNumber, err))
		}

		payloadMap := make(map[string]any, len(headers))
		for i, header := range headers {
			value := ""
			if i < len(record) {
				value = strings.TrimSpace(record[i])
			}
			payloadMap[header] = value
		}

		payloadBytes, err := json.Marshal(payloadMap)
		if err != nil {
			return failBatch(fmt.Errorf("failed to marshal row %d payload: %w", rowNumber, err))
		}

		_, err = c.importRepository.CreateImportRawRow(ctx, db.CreateImportRawRowParams{
			BatchID:   batch.ID,
			RowNumber: int32(rowNumber),
			Payload:   payloadBytes,
		})
		if err != nil {
			return failBatch(fmt.Errorf("failed to insert raw row %d: %w", rowNumber, err))
		}

		rowCount++
		rowNumber++

		if rowCount%100 == 0 {
			progress := calculateProgress(rowCount)
			msg = fmt.Sprintf("Imported %d rows", rowCount)
			_ = c.processRepository.UpdateProgress(ctx, p.ID, progress, &msg)
		}
	}

	_, err = c.importRepository.UpdateImportBatchStatus(ctx, db.UpdateImportBatchStatusParams{
		ID:     batch.ID,
		Status: "PENDING",
		Notes: sql.NullString{
			String: fmt.Sprintf("Uploaded %d raw rows. Awaiting batch processing.", rowCount),
			Valid:  true,
		},
	})
	if err != nil {
		return failBatch(fmt.Errorf("failed to mark batch pending: %w", err))
	}

	_, err = json.Marshal(SurveillanceBatchProcessPayload{
		BatchID: batch.ID,
	})
	if err != nil {
		return failBatch(fmt.Errorf("failed to marshal batch process payload: %w", err))
	}

	_, err = c.processRepository.CreateProcess(ctx, db.CreateProcessParams{
		DocumentID:  p.DocumentID,
		ProcessType: string(model.ProcessTypeSurveillanceBatchProcess),
		CreatedBy:   p.CreatedBy,
	})
	if err != nil {
		return failBatch(fmt.Errorf("failed to queue batch processing: %w", err))
	}

	msg = fmt.Sprintf("Imported %d raw rows successfully", rowCount)
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 90, &msg)

	msg = "Raw import completed"
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 100, &msg)

	return nil
}

func normalizeHeader(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	return value
}

func inferDatasetType(headers []string) string {
	headerSet := make(map[string]struct{}, len(headers))
	for _, h := range headers {
		headerSet[h] = struct{}{}
	}

	hasAny := func(keys ...string) bool {
		for _, key := range keys {
			if _, ok := headerSet[key]; ok {
				return true
			}
		}
		return false
	}

	hasFacility := hasAny("facility", "facility_name")
	hasDisease := hasAny("disease", "disease_name")
	hasValue := hasAny("value", "cases", "count", "metric_value")
	hasRegion := hasAny("region", "region_name")
	hasDistrict := hasAny("district", "district_name")
	hasWeek := hasAny("week", "weeks", "epi_week", "epiweek")
	hasMaroon := hasAny("maroon")
	hasRed := hasAny("red")
	hasYellow := hasAny("yellow")
	hasGreen := hasAny("green")
	hasNational := hasAny("national", "country", "uganda")

	switch {
	case hasFacility && hasDisease && hasValue:
		return "facility_metrics"
	case hasDistrict && hasMaroon && hasRed && hasYellow && hasGreen:
		return "district_status"
	case hasRegion && hasDisease && hasWeek:
		return "region_status"
	case hasNational && hasDisease && hasWeek:
		return "national_status"
	default:
		return "unknown"
	}
}

func calculateProgress(rowCount int) int32 {
	progress := int32(20 + (rowCount / 100))
	if progress > 85 {
		return 85
	}
	return progress
}
