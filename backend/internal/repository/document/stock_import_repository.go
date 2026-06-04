package document

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/moh-sso-dashboard/internal/model"
)

type StockImportRepository interface {
	InsertNMSStockIssue(ctx context.Context, tx *sql.Tx, row model.NMSStockIssue) error
	UpsertNMSStockIssuesBatch(ctx context.Context, tx *sql.Tx, rows []model.NMSStockIssue) error
	SoftDeleteNMSStockIssues(ctx context.Context, tx *sql.Tx, keepHashes []string) error
	InvalidateNMSStockIssuesByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error

	InsertNMSStockOnHand(ctx context.Context, tx *sql.Tx, row model.NMSStockOnHand) error
	UpsertNMSStockOnHandBatch(ctx context.Context, tx *sql.Tx, rows []model.NMSStockOnHand) error
	SoftDeleteNMSStockOnHand(ctx context.Context, tx *sql.Tx, keepHashes []string) error
	InvalidateNMSStockOnHandByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error

	InsertJMSStockIssue(ctx context.Context, tx *sql.Tx, row model.JMSStockIssue) error
	UpsertJMSStockIssuesBatch(ctx context.Context, tx *sql.Tx, rows []model.JMSStockIssue) error
	SoftDeleteJMSStockIssues(ctx context.Context, tx *sql.Tx, keepHashes []string) error
	InvalidateJMSStockIssuesByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error

	InsertJMSStockOnHand(ctx context.Context, tx *sql.Tx, row model.JMSStockOnHand) error
	UpsertJMSStockOnHandBatch(ctx context.Context, tx *sql.Tx, rows []model.JMSStockOnHand) error
	SoftDeleteJMSStockOnHand(ctx context.Context, tx *sql.Tx, keepHashes []string) error
	InvalidateJMSStockOnHandByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error

	// ── GF Pipeline ──────────────────────────────────────────────────────────────
	UpsertGFPipelineBatch(ctx context.Context, tx *sql.Tx, rows []model.GFPipelineRow) error
	SoftDeleteGFPipeline(ctx context.Context, tx *sql.Tx, keepHashes []string) error
	InvalidateGFPipelineByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error

	// ── GDF TB Orders ─────────────────────────────────────────────────────────────
	UpsertGDFTBOrdersBatch(ctx context.Context, tx *sql.Tx, rows []model.GDFTBOrder) error
	SoftDeleteGDFTBOrders(ctx context.Context, tx *sql.Tx, keepHashes []string) error
	InvalidateGDFTBOrdersByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error

	// ── GHSC-PSM ──────────────────────────────────────────────────────────────────
	UpsertGHSCPSMLabBatch(ctx context.Context, tx *sql.Tx, rows []model.GHSCPSMLab) error
	SoftDeleteGHSCPSMLab(ctx context.Context, tx *sql.Tx, keepHashes []string) error
	InvalidateGHSCPSMLabByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error

	UpsertGHSCPSMPharmaBatch(ctx context.Context, tx *sql.Tx, rows []model.GHSCPSMCommodity) error
	SoftDeleteGHSCPSMPharma(ctx context.Context, tx *sql.Tx, keepHashes []string) error
	InvalidateGHSCPSMPharmaByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error

	UpsertGHSCPSMMalariaBatch(ctx context.Context, tx *sql.Tx, rows []model.GHSCPSMCommodity) error
	SoftDeleteGHSCPSMMalaria(ctx context.Context, tx *sql.Tx, keepHashes []string) error
	InvalidateGHSCPSMMalariaByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error
}
