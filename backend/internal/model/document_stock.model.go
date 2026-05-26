package model

import (
	"time"

	"github.com/google/uuid"
)

type NMSStockIssue struct {
	DocumentID uuid.UUID
	RowNumber  int

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
