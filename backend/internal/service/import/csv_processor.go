package service

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	documentRepository "github.com/moh-sso-dashboard/internal/features/documents"
	processRepository "github.com/moh-sso-dashboard/internal/repository/processes"
	"github.com/moh-sso-dashboard/internal/storage"
)

type CSVProcessor struct {
	documentRepository documentRepository.DocumentRepository
	processRepository  processRepository.ProcessRepository
	fileRepository     documentRepository.FileRepository
	storage            storage.Storage
	remoteDB           *sql.DB
}

func NewCSVProcessor(documentRepository documentRepository.DocumentRepository, processRepository processRepository.ProcessRepository, fileRepository documentRepository.FileRepository, storage storage.Storage, remoteDB *sql.DB) *CSVProcessor {
	return &CSVProcessor{
		documentRepository: documentRepository,
		processRepository:  processRepository,
		fileRepository:     fileRepository,
		storage:            storage,
		remoteDB:           remoteDB,
	}
}

func (c *CSVProcessor) Process(
	ctx context.Context,
	p db.Process,
) error {
	// 1. Get document
	doc, err := c.documentRepository.GetDocument(ctx, p.DocumentID)
	if err != nil {
		return err
	}

	templateCode, err := getTemplateCodeOrAdHoc(doc)
	if err != nil {
		return fmt.Errorf("csv processor: %w", err)
	}
	templateCode = strings.ToUpper(strings.TrimSpace(templateCode))
	isAdHoc := templateCode == adHocTemplateCode

	reportDate := getOptionalReportDateFromDocument(doc)

	replaceDocumentID, err := getReplaceDocumentIDFromDocument(doc)
	if err != nil {
		return fmt.Errorf("csv processor: %w", err)
	}

	const sheetCode = "default"

	msg := "Opening file"
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 5, &msg)

	// 2. Download file
	fileReader, err := c.storage.Download(ctx, doc.ObjectKey)
	if err != nil {
		return err
	}
	defer fileReader.Close()

	msg = "Parsing CSV"
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 10, &msg)

	// 3. CSV reader
	reader := csv.NewReader(fileReader)
	reader.TrimLeadingSpace = true
	reader.ReuseRecord = true

	headers, err := reader.Read()
	if err != nil {
		return err
	}

	// Normalize headers once
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}

	// 4. If replacing a previous upload, invalidate its rows and file record first.
	if replaceDocumentID != nil || isAdHoc {
		invalidateID := doc.ID
		if replaceDocumentID != nil {
			invalidateID = *replaceDocumentID
		}
		tx, err := c.remoteDB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}
		if err := c.fileRepository.InvalidateByDocument(ctx, tx, invalidateID); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("invalidate previous document rows: %w", err)
		}
		if err := c.fileRepository.InvalidateFile(ctx, tx, invalidateID); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("invalidate previous file record: %w", err)
		}
		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// 5. Create file record
	var fileKey int64
	{
		tx, err := c.remoteDB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		fileKey, err = c.fileRepository.CreateFile(
			ctx,
			tx,
			doc.ID,
			templateCode,
			doc.OriginalFilename,
			doc.ObjectKey,
			reportDate,
		)
		if err != nil {
			_ = tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// 6. Batch upsert rows, collecting all hashes for soft-delete
	const batchSize = 2000
	rowCount := 0
	lastProgressUpdate := time.Now()

	batch := make([]documentRepository.CustomImportRow, 0, batchSize)
	allHashes := make([]string, 0, 10000)

	flushBatch := func() error {
		if len(batch) == 0 {
			return nil
		}

		tx, err := c.remoteDB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		if err := c.fileRepository.UpsertRowsBatch(ctx, tx, batch); err != nil {
			_ = tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			return err
		}

		batch = batch[:0]
		return nil
	}

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		rowJSON := make(map[string]string, len(headers))
		for i, header := range headers {
			if i < len(record) {
				rowJSON[header] = record[i]
			} else {
				rowJSON[header] = ""
			}
		}

		rowData := make(map[string]any, len(rowJSON))
		for k, v := range rowJSON {
			rowData[k] = v
		}

		rowCount++
		hash := hashRawData(doc.ID, sheetCode, rowCount, rowData)

		batch = append(batch, documentRepository.CustomImportRow{
			FileKey:      fileKey,
			DocumentID:   doc.ID,
			TemplateCode: templateCode,
			SheetCode:    sheetCode,
			RowNumber:    rowCount,
			ReportDate:   reportDate,
			Data:         rowData,
			RowHash:      hash,
		})
		allHashes = append(allHashes, hash)

		if len(batch) >= batchSize {
			if err := flushBatch(); err != nil {
				return err
			}
		}

		if rowCount%5000 == 0 || time.Since(lastProgressUpdate) > 3*time.Second {
			progress := int32(10 + (rowCount / 1000))
			if progress > 85 {
				progress = 85
			}
			msg := fmt.Sprintf("Processed %d rows", rowCount)
			_ = c.processRepository.UpdateProgress(ctx, p.ID, progress, &msg)
			lastProgressUpdate = time.Now()
		}
	}

	if err := flushBatch(); err != nil {
		return err
	}

	// 7. Soft-delete rows from previous uploads that are no longer in this file.
	if !isAdHoc {
		tx, err := c.remoteDB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		if err := c.fileRepository.SoftDeleteBySheet(ctx, tx, doc.ID, templateCode, sheetCode, allHashes); err != nil {
			_ = tx.Rollback()
			return err
		}

		if err := tx.Commit(); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	msg = fmt.Sprintf("Upserted %d rows successfully", rowCount)
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 90, &msg)

	msg = "Finished processing"
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 100, &msg)

	return nil
}
