package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	db "github.com/moh-sso-dashboard/internal/db/sqlc"
	documentRepository "github.com/moh-sso-dashboard/internal/repository/document"
	processRepository "github.com/moh-sso-dashboard/internal/repository/processes"
	"github.com/moh-sso-dashboard/internal/storage"
)

type CSVProcessor struct {
	documentRepository documentRepository.DocumentRepository
	processRepository  processRepository.ProcessRepository
	fileRepository     documentRepository.FileRepository
	storage            storage.Storage
	db                 db.Store
}

func NewCSVProcessor(documentRepository documentRepository.DocumentRepository, processRepository processRepository.ProcessRepository, storage storage.Storage, db db.Store) *CSVProcessor {
	return &CSVProcessor{
		documentRepository: documentRepository,
		processRepository:  processRepository,
		storage:            storage,
		db:                 db}
}

func (c *CSVProcessor) Process(
	ctx context.Context,
	p db.Process,
) error {

	// --------------------------------------------------
	// 1️⃣ Get Document
	// --------------------------------------------------
	doc, err := c.documentRepository.GetDocument(ctx, p.DocumentID)
	if err != nil {
		return err
	}

	msg := "Opening file"
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 5, &msg)

	// --------------------------------------------------
	// 2️⃣ Download file
	// --------------------------------------------------
	fileReader, err := c.storage.Download(ctx, doc.ObjectKey)
	if err != nil {
		return err
	}
	defer fileReader.Close()

	msg = "Parsing CSV"
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 10, &msg)

	// --------------------------------------------------
	// 3️⃣ CSV Reader
	// --------------------------------------------------
	reader := csv.NewReader(fileReader)
	reader.TrimLeadingSpace = true

	headers, err := reader.Read()
	if err != nil {
		return err
	}

	// --------------------------------------------------
	// 4️⃣ Execute Transaction Properly
	// --------------------------------------------------
	err = c.db.ExecTx(ctx, func(q db.DBTX) error {

		// 5️⃣ Create File Record
		fileKey, err := c.fileRepository.CreateCustomFile(
			ctx,
			q, // pass Querier instead of *sql.Tx
			doc.OriginalFilename,
			doc.ObjectKey,
		)
		if err != nil {
			return err
		}

		// 6️⃣ Stream + Insert
		rowCount := 0
		batchSize := 500

		for {
			record, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			rowJSON := make(map[string]interface{})

			for i, value := range record {
				if i < len(headers) {
					rowJSON[headers[i]] = value
				}
			}

			jsonBytes, err := json.Marshal(rowJSON)
			if err != nil {
				return err
			}

			if err := c.fileRepository.InsertCustomData(
				ctx,
				q, // pass Querier
				fileKey,
				jsonBytes,
			); err != nil {
				return err
			}

			rowCount++

			if rowCount%batchSize == 0 {
				progress := int32(10 + (rowCount / 100))
				if progress > 85 {
					progress = 85
				}
				msg := fmt.Sprintf("Processed %d rows", rowCount)
				_ = c.processRepository.UpdateProgress(ctx, p.ID, progress, &msg)
			}
		}

		msg := fmt.Sprintf("Inserted rows successfully")
		_ = c.processRepository.UpdateProgress(ctx, p.ID, 90, &msg)

		return nil
	})

	if err != nil {
		return err
	}

	msg = "Finished processing"
	_ = c.processRepository.UpdateProgress(ctx, p.ID, 100, &msg)

	return nil
}
