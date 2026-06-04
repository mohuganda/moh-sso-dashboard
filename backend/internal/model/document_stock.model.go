package model

import (
	"time"

	"github.com/google/uuid"
)

type NMSStockIssue struct {
	DocumentID uuid.UUID
	RowNumber  int
	RowHash    string
	ReportDate *time.Time

	OrderType          string
	ShipToFacilityCode string
	ShipToFacilityName string
	BillToFacilityCode string
	BillToFacilityName string
	District           string
	LOC                string
	FundingSource      string
	Cycle              string
	ItemCode           string
	ItemDescription    string
	UnitOfMeasure      string
	OrderNumber        string
	OrderQuantity      *float64
	LineID             string
	QuantityShipped    *float64
	LotNumber          string
	LotExpirationDate  *time.Time
	ListPrice          *float64
	SellingPrice       *float64
	CurrencyCode       string
	AmountUGX          *float64
	ShipToLocation     string
	ShipConfirmDate    *time.Time
	ShipConfirmedBy    string

	RawPayload map[string]any
}

type NMSStockOnHand struct {
	DocumentID uuid.UUID
	RowNumber  int
	RowHash    string
	ReportDate *time.Time

	ItemCode            string
	ItemDescription     string
	ItemInventoryStatus string
	FundingSource       string
	SubInventory        string
	Locator             string
	LotNumber           string
	ExpiryDate          *time.Time
	PalletID            string
	UOM                 string
	OnHandQuantity      *float64

	RawPayload map[string]any
}

type JMSStockIssue struct {
	DocumentID uuid.UUID
	RowNumber  int
	RowHash    string
	ReportDate *time.Time

	OrderNo            string
	SellToCustomerNo   string
	CustomerName       string
	District           string
	Reference          string
	PartNo             string
	PartDescription    string
	OwningCustomerNo   string
	OwningCustomerName string
	UOM                string
	ItemStatus         string
	SoldQty            *float64
	DesiredQty         *float64
	UnitPrice          *float64
	DateSold           *time.Time
	AssociationNo      string
	ExpiryDates        string
	OFR                *float64

	RawPayload map[string]any
}

type JMSStockOnHand struct {
	DocumentID uuid.UUID
	RowNumber  int
	RowHash    string
	ReportDate *time.Time

	LocationCode       string
	LocationName       string
	BinCode            string
	ReceiptDate        *time.Time
	ReceiptNo          string
	PurchaseOrderNo    string
	No                 string
	Description        string
	LotNo              string
	ExpirationDate     *time.Time
	AvailableQuantity  *float64
	UnitOfMeasureCode  string
	UnitCost           *float64
	Ownership          string
	TotalCost          *float64
	TransferOrderNo    string
	FromLocation       string
	ToLocation         string
	PostingDescription string
	YourReference      string

	RawPayload map[string]any
}

// ── Global Fund Pipeline TrackAndTrace ────────────────────────────────────────

type GFPipelineRow struct {
	DocumentID              uuid.UUID
	RowNumber               int
	RowHash                 string
	ReportDate              *time.Time

	PSAName                 string
	Country                 string
	GrantName               string
	GrantBudgetID           string
	ReqNumber               string
	EPONumber               string
	PONumber                string
	ShipmentNumber          string
	Status                  string
	VendorGroupName         string
	ItemNameTGF             string
	SQSOItemQuantity        *float64
	INCOTermClient          string
	ShipmentMode            string
	CountryShipToCity       string
	POItemQuantity          *float64
	ShipmentItemQtyOrdered  *float64
	ConfirmedReceivedQty    *float64
	NumberOfPallets         *float64
	ShipmentGrossWeight     *float64
	ShipmentVolume          *float64
	NumberOfContainersTotal *float64
	EstimatedVendorIncoDate *time.Time
	EstimatedDeliveryDate   *time.Time
	DeliveryDateActual      *time.Time
	SOItemLinenumber        string

	RawPayload map[string]any
}

// ── GDF TB Orders ─────────────────────────────────────────────────────────────

type GDFTBOrder struct {
	DocumentID        uuid.UUID
	RowNumber         int
	RowHash           string
	ReportDate        *time.Time

	YearCreated       *int
	Country           string
	Line              string
	SerialNumber      string
	TotalCost         *float64
	OrderStatus       string
	ShipmentCode      string
	ProductCode       string
	INNCode           string
	Supplier          string
	QuantityShipped   *float64
	UnitsShipped      *float64
	Price             *float64
	EstimatedDelivery *time.Time
	ActualDelivery    *time.Time
	IncoTerm          string
	ShippingMode      string

	RawPayload map[string]any
}

// ── GHSC-PSM Lab ──────────────────────────────────────────────────────────────

type GHSCPSMLab struct {
	DocumentID            uuid.UUID
	RowNumber             int
	RowHash               string
	ReportDate            *time.Time

	LineNumber            string
	CommodityCategory     string
	ItemDescription       string
	UOM                   string
	StockOnHand           *float64
	AMC                   *float64
	MOS                   *float64
	QuantityOnOrder       *float64
	RequestedDeliveryDate *time.Time
	EstimatedDeliveryDate *time.Time
	Status                string
	Comment               string

	RawPayload map[string]any
}

// ── GHSC-PSM Pharma / Malaria (same structure, separate tables) ───────────────

type GHSCPSMCommodity struct {
	DocumentID            uuid.UUID
	RowNumber             int
	RowHash               string
	ReportDate            *time.Time

	LineNumber            string
	CommodityCategory     string
	ItemDescription       string
	UOM                   string
	StockOnHand           *float64
	AMC                   *float64
	MOS                   *float64
	QuantityOnOrder       *float64
	EstimatedDeliveryDate *time.Time
	Status                string

	RawPayload map[string]any
}
