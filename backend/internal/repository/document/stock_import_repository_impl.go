package document

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/moh-sso-dashboard/internal/model"
)

type stockImportRepository struct{}

func NewStockImportRepository() StockImportRepository {
	return &stockImportRepository{}
}

func toJSONB(payload map[string]any) ([]byte, error) {
	if payload == nil {
		payload = map[string]any{}
	}
	return json.Marshal(payload)
}

func placeholders(rowIndex, colsPerRow int) string {
	offset := rowIndex * colsPerRow
	parts := make([]string, colsPerRow)
	for i := 0; i < colsPerRow; i++ {
		parts[i] = fmt.Sprintf("$%d", offset+i+1)
	}
	return "(" + strings.Join(parts, ",") + ")"
}

func softDelete(ctx context.Context, tx *sql.Tx, table string, keepHashes []string) error {
	_, err := tx.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s
		SET is_valid = FALSE, last_updated = NOW()
		WHERE is_valid = TRUE
		  AND row_hash IS NOT NULL
		  AND NOT (row_hash = ANY($1))
	`, table), pq.Array(keepHashes))
	return err
}

// ── NMS Stock Issues ──────────────────────────────────────────────────────────

func (r *stockImportRepository) InsertNMSStockIssue(ctx context.Context, tx *sql.Tx, row model.NMSStockIssue) error {
	return r.UpsertNMSStockIssuesBatch(ctx, tx, []model.NMSStockIssue{row})
}

func (r *stockImportRepository) UpsertNMSStockIssuesBatch(
	ctx context.Context,
	tx *sql.Tx,
	rows []model.NMSStockIssue,
) error {
	if len(rows) == 0 {
		return nil
	}

	const colsPerRow = 30

	args := make([]any, 0, len(rows)*colsPerRow)
	valueStrings := make([]string, 0, len(rows))

	for i, row := range rows {
		raw, err := toJSONB(row.RawPayload)
		if err != nil {
			return err
		}
		valueStrings = append(valueStrings, placeholders(i, colsPerRow))
		args = append(args,
			row.DocumentID, row.RowNumber,
			row.OrderType, row.ShipToFacilityCode, row.ShipToFacilityName,
			row.BillToFacilityCode, row.BillToFacilityName, row.District, row.LOC,
			row.FundingSource, row.Cycle, row.ItemCode, row.ItemDescription,
			row.UnitOfMeasure, row.OrderNumber, row.OrderQuantity, row.LineID,
			row.QuantityShipped, row.LotNumber, row.LotExpirationDate,
			row.ListPrice, row.SellingPrice, row.CurrencyCode, row.AmountUGX,
			row.ShipToLocation, row.ShipConfirmDate, row.ShipConfirmedBy,
			raw, row.RowHash, row.ReportDate,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO import.nms_stock_issues (
			document_id, row_number,
			order_type, ship_to_facility_code, ship_to_facility_name,
			bill_to_facility_code, bill_to_facility_name, district, loc,
			funding_source, cycle, item_code, item_description,
			unit_of_measure, order_number, order_quantity, line_id,
			quantity_shipped, lot_number, lot_expiration_date,
			list_price, selling_price, currency_code, amount_ugx,
			ship_to_location, ship_confirm_date, ship_confirmed_by,
			raw_payload, row_hash, report_date
		) VALUES %s
		ON CONFLICT (row_hash) DO UPDATE SET
			document_id          = EXCLUDED.document_id,
			row_number           = EXCLUDED.row_number,
			order_type           = EXCLUDED.order_type,
			ship_to_facility_code = EXCLUDED.ship_to_facility_code,
			ship_to_facility_name = EXCLUDED.ship_to_facility_name,
			bill_to_facility_code = EXCLUDED.bill_to_facility_code,
			bill_to_facility_name = EXCLUDED.bill_to_facility_name,
			district             = EXCLUDED.district,
			loc                  = EXCLUDED.loc,
			funding_source       = EXCLUDED.funding_source,
			cycle                = EXCLUDED.cycle,
			item_code            = EXCLUDED.item_code,
			item_description     = EXCLUDED.item_description,
			unit_of_measure      = EXCLUDED.unit_of_measure,
			order_number         = EXCLUDED.order_number,
			order_quantity       = EXCLUDED.order_quantity,
			line_id              = EXCLUDED.line_id,
			quantity_shipped     = EXCLUDED.quantity_shipped,
			lot_number           = EXCLUDED.lot_number,
			lot_expiration_date  = EXCLUDED.lot_expiration_date,
			list_price           = EXCLUDED.list_price,
			selling_price        = EXCLUDED.selling_price,
			currency_code        = EXCLUDED.currency_code,
			amount_ugx           = EXCLUDED.amount_ugx,
			ship_to_location     = EXCLUDED.ship_to_location,
			ship_confirm_date    = EXCLUDED.ship_confirm_date,
			ship_confirmed_by    = EXCLUDED.ship_confirmed_by,
			raw_payload          = EXCLUDED.raw_payload,
			report_date          = EXCLUDED.report_date,
			last_updated         = NOW(),
			is_valid             = TRUE
	`, strings.Join(valueStrings, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *stockImportRepository) SoftDeleteNMSStockIssues(ctx context.Context, tx *sql.Tx, keepHashes []string) error {
	return softDelete(ctx, tx, "import.nms_stock_issues", keepHashes)
}

// ── NMS Stock On Hand ─────────────────────────────────────────────────────────

func (r *stockImportRepository) InsertNMSStockOnHand(ctx context.Context, tx *sql.Tx, row model.NMSStockOnHand) error {
	return r.UpsertNMSStockOnHandBatch(ctx, tx, []model.NMSStockOnHand{row})
}

func (r *stockImportRepository) UpsertNMSStockOnHandBatch(
	ctx context.Context,
	tx *sql.Tx,
	rows []model.NMSStockOnHand,
) error {
	if len(rows) == 0 {
		return nil
	}

	const colsPerRow = 16

	args := make([]any, 0, len(rows)*colsPerRow)
	valueStrings := make([]string, 0, len(rows))

	for i, row := range rows {
		raw, err := toJSONB(row.RawPayload)
		if err != nil {
			return err
		}
		valueStrings = append(valueStrings, placeholders(i, colsPerRow))
		args = append(args,
			row.DocumentID, row.RowNumber,
			row.ItemCode, row.ItemDescription, row.ItemInventoryStatus,
			row.FundingSource, row.SubInventory, row.Locator,
			row.LotNumber, row.ExpiryDate, row.PalletID, row.UOM,
			row.OnHandQuantity,
			raw, row.RowHash, row.ReportDate,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO import.nms_stock_on_hand (
			document_id, row_number,
			item_code, item_description, item_inventory_status,
			funding_source, subinventory, locator,
			lot_number, expiry_date, pallet_id, uom,
			on_hand_quantity,
			raw_payload, row_hash, report_date
		) VALUES %s
		ON CONFLICT (row_hash) DO UPDATE SET
			document_id          = EXCLUDED.document_id,
			row_number           = EXCLUDED.row_number,
			item_code            = EXCLUDED.item_code,
			item_description     = EXCLUDED.item_description,
			item_inventory_status = EXCLUDED.item_inventory_status,
			funding_source       = EXCLUDED.funding_source,
			subinventory         = EXCLUDED.subinventory,
			locator              = EXCLUDED.locator,
			lot_number           = EXCLUDED.lot_number,
			expiry_date          = EXCLUDED.expiry_date,
			pallet_id            = EXCLUDED.pallet_id,
			uom                  = EXCLUDED.uom,
			on_hand_quantity     = EXCLUDED.on_hand_quantity,
			raw_payload          = EXCLUDED.raw_payload,
			report_date          = EXCLUDED.report_date,
			last_updated         = NOW(),
			is_valid             = TRUE
	`, strings.Join(valueStrings, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *stockImportRepository) SoftDeleteNMSStockOnHand(ctx context.Context, tx *sql.Tx, keepHashes []string) error {
	return softDelete(ctx, tx, "import.nms_stock_on_hand", keepHashes)
}

// ── JMS Stock Issues ──────────────────────────────────────────────────────────

func (r *stockImportRepository) InsertJMSStockIssue(ctx context.Context, tx *sql.Tx, row model.JMSStockIssue) error {
	return r.UpsertJMSStockIssuesBatch(ctx, tx, []model.JMSStockIssue{row})
}

func (r *stockImportRepository) UpsertJMSStockIssuesBatch(
	ctx context.Context,
	tx *sql.Tx,
	rows []model.JMSStockIssue,
) error {
	if len(rows) == 0 {
		return nil
	}

	const colsPerRow = 23

	args := make([]any, 0, len(rows)*colsPerRow)
	valueStrings := make([]string, 0, len(rows))

	for i, row := range rows {
		raw, err := toJSONB(row.RawPayload)
		if err != nil {
			return err
		}
		valueStrings = append(valueStrings, placeholders(i, colsPerRow))
		args = append(args,
			row.DocumentID, row.RowNumber,
			row.OrderNo, row.SellToCustomerNo, row.CustomerName,
			row.District, row.Reference, row.PartNo, row.PartDescription,
			row.OwningCustomerNo, row.OwningCustomerName,
			row.UOM, row.ItemStatus,
			row.SoldQty, row.DesiredQty, row.UnitPrice,
			row.DateSold, row.AssociationNo, row.ExpiryDates, row.OFR,
			raw, row.RowHash, row.ReportDate,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO import.jms_stock_issues (
			document_id, row_number,
			order_no, sell_to_customer_no, customer_name,
			district, reference, part_no, part_description,
			owning_customer_no, owning_customer_name,
			uom, item_status,
			sold_qty, desired_qty, unit_price,
			date_sold, association_no, expiry_dates, ofr,
			raw_payload, row_hash, report_date
		) VALUES %s
		ON CONFLICT (row_hash) DO UPDATE SET
			document_id         = EXCLUDED.document_id,
			row_number          = EXCLUDED.row_number,
			order_no            = EXCLUDED.order_no,
			sell_to_customer_no = EXCLUDED.sell_to_customer_no,
			customer_name       = EXCLUDED.customer_name,
			district            = EXCLUDED.district,
			reference           = EXCLUDED.reference,
			part_no             = EXCLUDED.part_no,
			part_description    = EXCLUDED.part_description,
			owning_customer_no  = EXCLUDED.owning_customer_no,
			owning_customer_name = EXCLUDED.owning_customer_name,
			uom                 = EXCLUDED.uom,
			item_status         = EXCLUDED.item_status,
			sold_qty            = EXCLUDED.sold_qty,
			desired_qty         = EXCLUDED.desired_qty,
			unit_price          = EXCLUDED.unit_price,
			date_sold           = EXCLUDED.date_sold,
			association_no      = EXCLUDED.association_no,
			expiry_dates        = EXCLUDED.expiry_dates,
			ofr                 = EXCLUDED.ofr,
			raw_payload         = EXCLUDED.raw_payload,
			report_date         = EXCLUDED.report_date,
			last_updated        = NOW(),
			is_valid            = TRUE
	`, strings.Join(valueStrings, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *stockImportRepository) SoftDeleteJMSStockIssues(ctx context.Context, tx *sql.Tx, keepHashes []string) error {
	return softDelete(ctx, tx, "import.jms_stock_issues", keepHashes)
}

// ── JMS Stock On Hand ─────────────────────────────────────────────────────────

func (r *stockImportRepository) InsertJMSStockOnHand(ctx context.Context, tx *sql.Tx, row model.JMSStockOnHand) error {
	return r.UpsertJMSStockOnHandBatch(ctx, tx, []model.JMSStockOnHand{row})
}

func (r *stockImportRepository) UpsertJMSStockOnHandBatch(
	ctx context.Context,
	tx *sql.Tx,
	rows []model.JMSStockOnHand,
) error {
	if len(rows) == 0 {
		return nil
	}

	const colsPerRow = 25

	args := make([]any, 0, len(rows)*colsPerRow)
	valueStrings := make([]string, 0, len(rows))

	for i, row := range rows {
		raw, err := toJSONB(row.RawPayload)
		if err != nil {
			return err
		}
		valueStrings = append(valueStrings, placeholders(i, colsPerRow))
		args = append(args,
			row.DocumentID, row.RowNumber,
			row.LocationCode, row.LocationName, row.BinCode,
			row.ReceiptDate, row.ReceiptNo, row.PurchaseOrderNo,
			row.No, row.Description, row.LotNo, row.ExpirationDate,
			row.AvailableQuantity, row.UnitOfMeasureCode, row.UnitCost,
			row.Ownership, row.TotalCost, row.TransferOrderNo,
			row.FromLocation, row.ToLocation, row.PostingDescription,
			row.YourReference,
			raw, row.RowHash, row.ReportDate,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO import.jms_stock_on_hand (
			document_id, row_number,
			location_code, location_name, bin_code,
			receipt_date, receipt_no, purchase_order_no,
			no, description, lot_no, expiration_date,
			available_quantity, unit_of_measure_code, unit_cost,
			ownership, total_cost, transfer_order_no,
			from_location, to_location, posting_description,
			your_reference,
			raw_payload, row_hash, report_date
		) VALUES %s
		ON CONFLICT (row_hash) DO UPDATE SET
			document_id         = EXCLUDED.document_id,
			row_number          = EXCLUDED.row_number,
			location_code       = EXCLUDED.location_code,
			location_name       = EXCLUDED.location_name,
			bin_code            = EXCLUDED.bin_code,
			receipt_date        = EXCLUDED.receipt_date,
			receipt_no          = EXCLUDED.receipt_no,
			purchase_order_no   = EXCLUDED.purchase_order_no,
			no                  = EXCLUDED.no,
			description         = EXCLUDED.description,
			lot_no              = EXCLUDED.lot_no,
			expiration_date     = EXCLUDED.expiration_date,
			available_quantity  = EXCLUDED.available_quantity,
			unit_of_measure_code = EXCLUDED.unit_of_measure_code,
			unit_cost           = EXCLUDED.unit_cost,
			ownership           = EXCLUDED.ownership,
			total_cost          = EXCLUDED.total_cost,
			transfer_order_no   = EXCLUDED.transfer_order_no,
			from_location       = EXCLUDED.from_location,
			to_location         = EXCLUDED.to_location,
			posting_description = EXCLUDED.posting_description,
			your_reference      = EXCLUDED.your_reference,
			raw_payload         = EXCLUDED.raw_payload,
			report_date         = EXCLUDED.report_date,
			last_updated        = NOW(),
			is_valid            = TRUE
	`, strings.Join(valueStrings, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *stockImportRepository) SoftDeleteJMSStockOnHand(ctx context.Context, tx *sql.Tx, keepHashes []string) error {
	return softDelete(ctx, tx, "import.jms_stock_on_hand", keepHashes)
}

// ── Invalidate by document (replace/re-upload) ────────────────────────────────

func invalidateByDocument(ctx context.Context, tx *sql.Tx, table string, documentID uuid.UUID) error {
	_, err := tx.ExecContext(ctx, fmt.Sprintf(`
		UPDATE %s SET is_valid = FALSE, last_updated = NOW()
		WHERE document_id = $1 AND is_valid = TRUE
	`, table), documentID)
	return err
}

func (r *stockImportRepository) InvalidateNMSStockIssuesByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error {
	return invalidateByDocument(ctx, tx, "import.nms_stock_issues", documentID)
}

func (r *stockImportRepository) InvalidateNMSStockOnHandByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error {
	return invalidateByDocument(ctx, tx, "import.nms_stock_on_hand", documentID)
}

func (r *stockImportRepository) InvalidateJMSStockIssuesByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error {
	return invalidateByDocument(ctx, tx, "import.jms_stock_issues", documentID)
}

func (r *stockImportRepository) InvalidateJMSStockOnHandByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error {
	return invalidateByDocument(ctx, tx, "import.jms_stock_on_hand", documentID)
}

// ── GF Pipeline TrackAndTrace ─────────────────────────────────────────────────

func (r *stockImportRepository) UpsertGFPipelineBatch(
	ctx context.Context,
	tx *sql.Tx,
	rows []model.GFPipelineRow,
) error {
	if len(rows) == 0 {
		return nil
	}

	const colsPerRow = 31

	args := make([]any, 0, len(rows)*colsPerRow)
	valueStrings := make([]string, 0, len(rows))

	for i, row := range rows {
		raw, err := toJSONB(row.RawPayload)
		if err != nil {
			return err
		}
		valueStrings = append(valueStrings, placeholders(i, colsPerRow))
		args = append(args,
			row.DocumentID, row.RowNumber, row.ReportDate,
			row.PSAName, row.Country, row.GrantName, row.GrantBudgetID,
			row.ReqNumber, row.EPONumber, row.PONumber, row.ShipmentNumber,
			row.Status, row.VendorGroupName, row.ItemNameTGF,
			row.SQSOItemQuantity, row.INCOTermClient, row.ShipmentMode, row.CountryShipToCity,
			row.POItemQuantity, row.ShipmentItemQtyOrdered, row.ConfirmedReceivedQty,
			row.NumberOfPallets, row.ShipmentGrossWeight, row.ShipmentVolume, row.NumberOfContainersTotal,
			row.EstimatedVendorIncoDate, row.EstimatedDeliveryDate, row.DeliveryDateActual,
			row.SOItemLinenumber,
			raw, row.RowHash,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO import.gf_pipeline (
			document_id, row_number, report_date,
			psa_name, country, grant_name, grant_budget_id,
			req_number, epo_number, po_number, shipment_number,
			status, vendor_group_name, item_name_tgf,
			sq_so_item_quantity, inco_term_client, shipment_mode, country_ship_to_city,
			po_item_quantity, shipment_item_qty_ordered, confirmed_received_quantity,
			number_of_pallets, shipment_gross_weight, shipment_volume, number_of_containers_total,
			estimated_vendor_inco_date, estimated_delivery_date, delivery_date_actual,
			so_item_linenumber,
			raw_payload, row_hash
		) VALUES %s
		ON CONFLICT (row_hash) DO UPDATE SET
			document_id                  = EXCLUDED.document_id,
			row_number                   = EXCLUDED.row_number,
			report_date                  = EXCLUDED.report_date,
			psa_name                     = EXCLUDED.psa_name,
			country                      = EXCLUDED.country,
			grant_name                   = EXCLUDED.grant_name,
			grant_budget_id              = EXCLUDED.grant_budget_id,
			req_number                   = EXCLUDED.req_number,
			epo_number                   = EXCLUDED.epo_number,
			po_number                    = EXCLUDED.po_number,
			shipment_number              = EXCLUDED.shipment_number,
			status                       = EXCLUDED.status,
			vendor_group_name            = EXCLUDED.vendor_group_name,
			item_name_tgf                = EXCLUDED.item_name_tgf,
			sq_so_item_quantity          = EXCLUDED.sq_so_item_quantity,
			inco_term_client             = EXCLUDED.inco_term_client,
			shipment_mode                = EXCLUDED.shipment_mode,
			country_ship_to_city         = EXCLUDED.country_ship_to_city,
			po_item_quantity             = EXCLUDED.po_item_quantity,
			shipment_item_qty_ordered    = EXCLUDED.shipment_item_qty_ordered,
			confirmed_received_quantity  = EXCLUDED.confirmed_received_quantity,
			number_of_pallets            = EXCLUDED.number_of_pallets,
			shipment_gross_weight        = EXCLUDED.shipment_gross_weight,
			shipment_volume              = EXCLUDED.shipment_volume,
			number_of_containers_total   = EXCLUDED.number_of_containers_total,
			estimated_vendor_inco_date   = EXCLUDED.estimated_vendor_inco_date,
			estimated_delivery_date      = EXCLUDED.estimated_delivery_date,
			delivery_date_actual         = EXCLUDED.delivery_date_actual,
			so_item_linenumber           = EXCLUDED.so_item_linenumber,
			raw_payload                  = EXCLUDED.raw_payload,
			last_updated                 = NOW(),
			is_valid                     = TRUE
	`, strings.Join(valueStrings, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *stockImportRepository) SoftDeleteGFPipeline(ctx context.Context, tx *sql.Tx, keepHashes []string) error {
	return softDelete(ctx, tx, "import.gf_pipeline", keepHashes)
}

func (r *stockImportRepository) InvalidateGFPipelineByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error {
	return invalidateByDocument(ctx, tx, "import.gf_pipeline", documentID)
}

// ── GDF TB Orders ─────────────────────────────────────────────────────────────

func (r *stockImportRepository) UpsertGDFTBOrdersBatch(
	ctx context.Context,
	tx *sql.Tx,
	rows []model.GDFTBOrder,
) error {
	if len(rows) == 0 {
		return nil
	}

	const colsPerRow = 22

	args := make([]any, 0, len(rows)*colsPerRow)
	valueStrings := make([]string, 0, len(rows))

	for i, row := range rows {
		raw, err := toJSONB(row.RawPayload)
		if err != nil {
			return err
		}
		valueStrings = append(valueStrings, placeholders(i, colsPerRow))
		args = append(args,
			row.DocumentID, row.RowNumber, row.ReportDate,
			row.YearCreated, row.Country, row.Line, row.SerialNumber,
			row.TotalCost, row.OrderStatus, row.ShipmentCode,
			row.ProductCode, row.INNCode, row.Supplier,
			row.QuantityShipped, row.UnitsShipped, row.Price,
			row.EstimatedDelivery, row.ActualDelivery,
			row.IncoTerm, row.ShippingMode,
			raw, row.RowHash,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO import.gdf_tb_orders (
			document_id, row_number, report_date,
			year_created, country, line, serial_number,
			total_cost, order_status, shipment_code,
			product_code, inn_code, supplier,
			quantity_shipped, units_shipped, price,
			estimated_delivery, actual_delivery,
			inco_term, shipping_mode,
			raw_payload, row_hash
		) VALUES %s
		ON CONFLICT (row_hash) DO UPDATE SET
			document_id        = EXCLUDED.document_id,
			row_number         = EXCLUDED.row_number,
			report_date        = EXCLUDED.report_date,
			year_created       = EXCLUDED.year_created,
			country            = EXCLUDED.country,
			line               = EXCLUDED.line,
			serial_number      = EXCLUDED.serial_number,
			total_cost         = EXCLUDED.total_cost,
			order_status       = EXCLUDED.order_status,
			shipment_code      = EXCLUDED.shipment_code,
			product_code       = EXCLUDED.product_code,
			inn_code           = EXCLUDED.inn_code,
			supplier           = EXCLUDED.supplier,
			quantity_shipped   = EXCLUDED.quantity_shipped,
			units_shipped      = EXCLUDED.units_shipped,
			price              = EXCLUDED.price,
			estimated_delivery = EXCLUDED.estimated_delivery,
			actual_delivery    = EXCLUDED.actual_delivery,
			inco_term          = EXCLUDED.inco_term,
			shipping_mode      = EXCLUDED.shipping_mode,
			raw_payload        = EXCLUDED.raw_payload,
			last_updated       = NOW(),
			is_valid           = TRUE
	`, strings.Join(valueStrings, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *stockImportRepository) SoftDeleteGDFTBOrders(ctx context.Context, tx *sql.Tx, keepHashes []string) error {
	return softDelete(ctx, tx, "import.gdf_tb_orders", keepHashes)
}

func (r *stockImportRepository) InvalidateGDFTBOrdersByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error {
	return invalidateByDocument(ctx, tx, "import.gdf_tb_orders", documentID)
}

// ── GHSC-PSM Lab ──────────────────────────────────────────────────────────────

func (r *stockImportRepository) UpsertGHSCPSMLabBatch(
	ctx context.Context,
	tx *sql.Tx,
	rows []model.GHSCPSMLab,
) error {
	if len(rows) == 0 {
		return nil
	}

	const colsPerRow = 17

	args := make([]any, 0, len(rows)*colsPerRow)
	valueStrings := make([]string, 0, len(rows))

	for i, row := range rows {
		raw, err := toJSONB(row.RawPayload)
		if err != nil {
			return err
		}
		valueStrings = append(valueStrings, placeholders(i, colsPerRow))
		args = append(args,
			row.DocumentID, row.RowNumber, row.ReportDate,
			row.LineNumber, row.CommodityCategory, row.ItemDescription, row.UOM,
			row.StockOnHand, row.AMC, row.MOS, row.QuantityOnOrder,
			row.RequestedDeliveryDate, row.EstimatedDeliveryDate,
			row.Status, row.Comment,
			raw, row.RowHash,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO import.ghsc_psm_lab (
			document_id, row_number, report_date,
			line_number, commodity_category, item_description, uom,
			stock_on_hand, amc, mos, quantity_on_order,
			requested_delivery_date, estimated_delivery_date,
			status, comment,
			raw_payload, row_hash
		) VALUES %s
		ON CONFLICT (row_hash) DO UPDATE SET
			document_id             = EXCLUDED.document_id,
			row_number              = EXCLUDED.row_number,
			report_date             = EXCLUDED.report_date,
			line_number             = EXCLUDED.line_number,
			commodity_category      = EXCLUDED.commodity_category,
			item_description        = EXCLUDED.item_description,
			uom                     = EXCLUDED.uom,
			stock_on_hand           = EXCLUDED.stock_on_hand,
			amc                     = EXCLUDED.amc,
			mos                     = EXCLUDED.mos,
			quantity_on_order       = EXCLUDED.quantity_on_order,
			requested_delivery_date = EXCLUDED.requested_delivery_date,
			estimated_delivery_date = EXCLUDED.estimated_delivery_date,
			status                  = EXCLUDED.status,
			comment                 = EXCLUDED.comment,
			raw_payload             = EXCLUDED.raw_payload,
			last_updated            = NOW(),
			is_valid                = TRUE
	`, strings.Join(valueStrings, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *stockImportRepository) SoftDeleteGHSCPSMLab(ctx context.Context, tx *sql.Tx, keepHashes []string) error {
	return softDelete(ctx, tx, "import.ghsc_psm_lab", keepHashes)
}

func (r *stockImportRepository) InvalidateGHSCPSMLabByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error {
	return invalidateByDocument(ctx, tx, "import.ghsc_psm_lab", documentID)
}

// ── GHSC-PSM Pharma ───────────────────────────────────────────────────────────

func (r *stockImportRepository) UpsertGHSCPSMPharmaBatch(
	ctx context.Context,
	tx *sql.Tx,
	rows []model.GHSCPSMCommodity,
) error {
	return upsertGHSCPSMCommodityBatch(ctx, tx, rows, "import.ghsc_psm_pharma")
}

func (r *stockImportRepository) SoftDeleteGHSCPSMPharma(ctx context.Context, tx *sql.Tx, keepHashes []string) error {
	return softDelete(ctx, tx, "import.ghsc_psm_pharma", keepHashes)
}

func (r *stockImportRepository) InvalidateGHSCPSMPharmaByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error {
	return invalidateByDocument(ctx, tx, "import.ghsc_psm_pharma", documentID)
}

// ── GHSC-PSM Malaria ──────────────────────────────────────────────────────────

func (r *stockImportRepository) UpsertGHSCPSMMalariaBatch(
	ctx context.Context,
	tx *sql.Tx,
	rows []model.GHSCPSMCommodity,
) error {
	return upsertGHSCPSMCommodityBatch(ctx, tx, rows, "import.ghsc_psm_malaria")
}

func (r *stockImportRepository) SoftDeleteGHSCPSMMalaria(ctx context.Context, tx *sql.Tx, keepHashes []string) error {
	return softDelete(ctx, tx, "import.ghsc_psm_malaria", keepHashes)
}

func (r *stockImportRepository) InvalidateGHSCPSMMalariaByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error {
	return invalidateByDocument(ctx, tx, "import.ghsc_psm_malaria", documentID)
}

// ── UNFPA Pipeline ────────────────────────────────────────────────────────────

func (r *stockImportRepository) UpsertUNFPAPipelineBatch(
	ctx context.Context,
	tx *sql.Tx,
	rows []model.UNFPAPipelineRow,
) error {
	if len(rows) == 0 {
		return nil
	}

	const colsPerRow = 29

	args := make([]any, 0, len(rows)*colsPerRow)
	valueStrings := make([]string, 0, len(rows))

	for i, row := range rows {
		raw, err := toJSONB(row.RawPayload)
		if err != nil {
			return err
		}
		valueStrings = append(valueStrings, placeholders(i, colsPerRow))
		args = append(args,
			row.DocumentID, row.RowNumber, row.ReportDate,
			row.RequisitionNo, row.Dept, row.ProductID, row.QuantumItemNumber,
			row.UOM, row.MOHUnits, row.DKTUnits, row.MSIUnits, row.PSIUnits, row.IPPFUnits,
			row.TotalUnits, row.TotalCost, row.UnitPrice,
			row.Vendor, row.ReqLineNo, row.PONumber, row.PODueDate,
			row.OrderLifeCycle, row.FundStatus, row.Status, row.ETA,
			row.Tranche, row.FundingYear, row.Period,
			raw, row.RowHash,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO import.unfpa_pipeline (
			document_id, row_number, report_date,
			requisition_no, dept, product_id, quantum_item_number,
			uom, moh_units, dkt_units, msi_units, psi_units, ippf_units,
			total_units, total_cost, unit_price,
			vendor, req_line_no, po_number, po_due_date,
			order_life_cycle, fund_status, status, eta,
			tranche, funding_year, period,
			raw_payload, row_hash
		) VALUES %s
		ON CONFLICT (row_hash) DO UPDATE SET
			document_id      = EXCLUDED.document_id,
			row_number       = EXCLUDED.row_number,
			report_date      = EXCLUDED.report_date,
			requisition_no   = EXCLUDED.requisition_no,
			dept             = EXCLUDED.dept,
			product_id       = EXCLUDED.product_id,
			quantum_item_number = EXCLUDED.quantum_item_number,
			uom              = EXCLUDED.uom,
			moh_units        = EXCLUDED.moh_units,
			dkt_units        = EXCLUDED.dkt_units,
			msi_units        = EXCLUDED.msi_units,
			psi_units        = EXCLUDED.psi_units,
			ippf_units       = EXCLUDED.ippf_units,
			total_units      = EXCLUDED.total_units,
			total_cost       = EXCLUDED.total_cost,
			unit_price       = EXCLUDED.unit_price,
			vendor           = EXCLUDED.vendor,
			req_line_no      = EXCLUDED.req_line_no,
			po_number        = EXCLUDED.po_number,
			po_due_date      = EXCLUDED.po_due_date,
			order_life_cycle = EXCLUDED.order_life_cycle,
			fund_status      = EXCLUDED.fund_status,
			status           = EXCLUDED.status,
			eta              = EXCLUDED.eta,
			tranche          = EXCLUDED.tranche,
			funding_year     = EXCLUDED.funding_year,
			period           = EXCLUDED.period,
			raw_payload      = EXCLUDED.raw_payload,
			last_updated     = NOW(),
			is_valid         = TRUE
	`, strings.Join(valueStrings, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *stockImportRepository) SoftDeleteUNFPAPipeline(ctx context.Context, tx *sql.Tx, keepHashes []string) error {
	return softDelete(ctx, tx, "import.unfpa_pipeline", keepHashes)
}

func (r *stockImportRepository) InvalidateUNFPAPipelineByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error {
	return invalidateByDocument(ctx, tx, "import.unfpa_pipeline", documentID)
}

// ── UNFPA NMS Pipeline ────────────────────────────────────────────────────────

func (r *stockImportRepository) UpsertUNFPANMSPipelineBatch(
	ctx context.Context,
	tx *sql.Tx,
	rows []model.UNFPANMSPipelineRow,
) error {
	if len(rows) == 0 {
		return nil
	}

	const colsPerRow = 17

	args := make([]any, 0, len(rows)*colsPerRow)
	valueStrings := make([]string, 0, len(rows))

	for i, row := range rows {
		raw, err := toJSONB(row.RawPayload)
		if err != nil {
			return err
		}
		valueStrings = append(valueStrings, placeholders(i, colsPerRow))
		args = append(args,
			row.DocumentID, row.RowNumber, row.ReportDate,
			row.Item, row.ItemID, row.ItemName, row.MOT, row.ETA,
			row.Quantity, row.Value, row.Supplier, row.PONumber,
			row.InProductionUntil, row.ETAAsPerOffer, row.Status,
			raw, row.RowHash,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO import.unfpa_nms_pipeline (
			document_id, row_number, report_date,
			item, item_id, item_name, mot, eta,
			quantity, value, supplier, po_number,
			in_production_until, eta_as_per_offer, status,
			raw_payload, row_hash
		) VALUES %s
		ON CONFLICT (row_hash) DO UPDATE SET
			document_id          = EXCLUDED.document_id,
			row_number           = EXCLUDED.row_number,
			report_date          = EXCLUDED.report_date,
			item                 = EXCLUDED.item,
			item_id              = EXCLUDED.item_id,
			item_name            = EXCLUDED.item_name,
			mot                  = EXCLUDED.mot,
			eta                  = EXCLUDED.eta,
			quantity             = EXCLUDED.quantity,
			value                = EXCLUDED.value,
			supplier             = EXCLUDED.supplier,
			po_number            = EXCLUDED.po_number,
			in_production_until  = EXCLUDED.in_production_until,
			eta_as_per_offer     = EXCLUDED.eta_as_per_offer,
			status               = EXCLUDED.status,
			raw_payload          = EXCLUDED.raw_payload,
			last_updated         = NOW(),
			is_valid             = TRUE
	`, strings.Join(valueStrings, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}

func (r *stockImportRepository) SoftDeleteUNFPANMSPipeline(ctx context.Context, tx *sql.Tx, keepHashes []string) error {
	return softDelete(ctx, tx, "import.unfpa_nms_pipeline", keepHashes)
}

func (r *stockImportRepository) InvalidateUNFPANMSPipelineByDocument(ctx context.Context, tx *sql.Tx, documentID uuid.UUID) error {
	return invalidateByDocument(ctx, tx, "import.unfpa_nms_pipeline", documentID)
}

func upsertGHSCPSMCommodityBatch(ctx context.Context, tx *sql.Tx, rows []model.GHSCPSMCommodity, table string) error {
	if len(rows) == 0 {
		return nil
	}

	const colsPerRow = 15

	args := make([]any, 0, len(rows)*colsPerRow)
	valueStrings := make([]string, 0, len(rows))

	for i, row := range rows {
		raw, err := toJSONB(row.RawPayload)
		if err != nil {
			return err
		}
		valueStrings = append(valueStrings, placeholders(i, colsPerRow))
		args = append(args,
			row.DocumentID, row.RowNumber, row.ReportDate,
			row.LineNumber, row.CommodityCategory, row.ItemDescription, row.UOM,
			row.StockOnHand, row.AMC, row.MOS, row.QuantityOnOrder,
			row.EstimatedDeliveryDate,
			row.Status,
			raw, row.RowHash,
		)
	}

	query := fmt.Sprintf(`
		INSERT INTO %s (
			document_id, row_number, report_date,
			line_number, commodity_category, item_description, uom,
			stock_on_hand, amc, mos, quantity_on_order,
			estimated_delivery_date,
			status,
			raw_payload, row_hash
		) VALUES %%s
		ON CONFLICT (row_hash) DO UPDATE SET
			document_id             = EXCLUDED.document_id,
			row_number              = EXCLUDED.row_number,
			report_date             = EXCLUDED.report_date,
			line_number             = EXCLUDED.line_number,
			commodity_category      = EXCLUDED.commodity_category,
			item_description        = EXCLUDED.item_description,
			uom                     = EXCLUDED.uom,
			stock_on_hand           = EXCLUDED.stock_on_hand,
			amc                     = EXCLUDED.amc,
			mos                     = EXCLUDED.mos,
			quantity_on_order       = EXCLUDED.quantity_on_order,
			estimated_delivery_date = EXCLUDED.estimated_delivery_date,
			status                  = EXCLUDED.status,
			raw_payload             = EXCLUDED.raw_payload,
			last_updated            = NOW(),
			is_valid                = TRUE
	`, table)

	query = fmt.Sprintf(query, strings.Join(valueStrings, ","))

	_, err := tx.ExecContext(ctx, query, args...)
	return err
}
