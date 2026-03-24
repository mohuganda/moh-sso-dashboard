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

	fileName := stringPtr(doc.OriginalFilename)
	importedBy := stringPtrFromUUID(p.CreatedBy)

	msg = "Creating import batch"
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 15, &msg)

	batch, err := c.importRepository.CreateImportBatch(ctx, db.CreateImportBatchParams{
		SourceName: "document_upload",
		FileName: sql.NullString{
			String: *fileName,
			Valid:  true,
		},
		DatasetType: inferDatasetType(headers),
		ImportedBy: sql.NullString{
			String: *importedBy,
			Valid:  true,
		},
		Status: "PROCESSING",
		Notes: sql.NullString{
			String: "Processing file",
			Valid:  true,
		},
	})
	if err != nil {
		return err
	}

	batchID = batch.ID
	batchCreated = true

	msg = "Importing rows"
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 20, &msg)

	rowCount := 0
	rowNumber := 1 // data row number, excluding header row

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
			var value string
			if i < len(record) {
				value = strings.TrimSpace(record[i])
			} else {
				value = ""
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
		Status: "COMPLETED",
		Notes: sql.NullString{
			String: "Successfully uploaded file",
			Valid:  true,
		},
	})
	if err != nil {
		return failBatch(fmt.Errorf("failed to mark batch completed: %w", err))
	}

	msg = fmt.Sprintf("Inserted %d rows successfully", rowCount)
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 90, &msg)

	msg = "Finished processing"
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

	_, hasFacility := headerSet["facility"]
	_, hasDisease := headerSet["disease"]
	_, hasValue := headerSet["value"]
	_, hasRegion := headerSet["region"]
	_, hasDistrict := headerSet["district"]
	_, hasWeek := headerSet["weeks"]
	_, hasMaroon := headerSet["maroon"]
	_, hasRed := headerSet["red"]
	_, hasYellow := headerSet["yellow"]
	_, hasGreen := headerSet["green"]

	switch {
	case hasFacility && hasDisease && hasValue:
		return "facility_metrics"
	case hasDistrict && hasMaroon && hasRed && hasYellow && hasGreen:
		return "district_status"
	case hasRegion && hasDisease && hasWeek:
		return "region_status"
	default:
		return "unknown"
	}
}

func calculateProgress(rowCount int) int32 {
	// keeps row import progress between 20 and 85
	progress := int32(20 + (rowCount / 100))
	if progress > 85 {
		return 85
	}
	return progress
}

func stringPtr(value string) *string {
	v := strings.TrimSpace(value)
	if v == "" {
		return nil
	}
	return &v
}

func stringPtrFromUUID(id uuid.UUID) *string {
	if id == uuid.Nil {
		return nil
	}
	v := id.String()
	return &v
}
