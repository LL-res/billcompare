package types

import "github.com/shopspring/decimal"

type BillItem struct {
	ProductType string
	ChargeType  string
	Cost        decimal.Decimal
	Arrears     decimal.Decimal
}
type DataSet struct {
	Data []Diff `json:"data"`
}
type Diff struct {
	ProductDetail       string          `json:"product_detail"`
	ChargeType          string          `json:"charge_type"`
	SchedulxCost        decimal.Decimal `json:"schedulx_cost"`
	AlibabaCloudCost    decimal.Decimal `json:"alibaba_cloud_cost"`
	SchedulxArrears     decimal.Decimal `json:"schedulx_arrears"`
	AlibabaCloudArrears decimal.Decimal `json:"alibaba_cloud_arrears"`
}
