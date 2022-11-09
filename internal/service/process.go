package service

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/galayx-future/billcompare/internal/constants"
	"github.com/galayx-future/billcompare/internal/types"
	"github.com/shopspring/decimal"
	"github.com/xuri/excelize/v2"
	"io"
	"log"
	"os"
	"strings"
)

type Processor struct {
	CloudBills map[string]*types.BillItem
	LocalBills map[string]*types.BillItem
}

func NewProcessor() *Processor {
	return &Processor{
		CloudBills: make(map[string]*types.BillItem),
		LocalBills: make(map[string]*types.BillItem),
	}
}
func (p *Processor) CSVLoadCloud(path string) error {
	fs, err := os.Open(path)
	if err != nil {
		log.Fatalf("!E %+v", err)
		return err
	}
	defer fs.Close()
	r := csv.NewReader(fs)
	isheader := true
	var headerIndex map[int]string
	indexs := make([]int, 0)
	for {
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("!E open csv fail %+v", err)
			continue
		}

		if isheader {
			headerIndex = getHeaderIndex(row)
			for k := range headerIndex {
				indexs = append(indexs, k)
			}
			isheader = false
			continue
		}

		keyTemp := make([]string, 2)
		var value types.BillItem
		for _, index := range indexs {
			switch headerIndex[index] {
			case constants.ProductType:
				keyTemp[0] = row[index]
				value.ProductType = row[index]
			case constants.ChargeType:
				keyTemp[1] = row[index]
				value.ChargeType = row[index]
			case constants.Cost:
				value.Cost, err = decimal.NewFromString(row[index])
				if err != nil {
					log.Printf("!E parse fail %+v", err)
				}
			case constants.Arrears:
				value.Arrears, err = decimal.NewFromString(row[index])
				if err != nil {
					log.Printf("!E parse fail %+v", err)
				}
			}
		}
		key := strings.Join(keyTemp, "-")
		if _, ok := p.CloudBills[key]; ok {
			p.CloudBills[key].Cost = p.CloudBills[key].Cost.Add(value.Cost)
			p.CloudBills[key].Arrears = p.CloudBills[key].Arrears.Add(value.Arrears)
		} else {
			p.CloudBills[key] = &value
		}

	}
	return nil
}
func (p *Processor) XLSXLoadLocal(path string) error {
	f, err := excelize.OpenFile(path)
	defer f.Close()
	if err != nil {
		log.Printf("!E open xlsx fail %+v", err)
		return err
	}
	rows, err := f.GetRows("SheetJS")
	rows[1][0] = "产品明细"
	rows[1][1] = "付费类型"
	if err != nil {
		log.Printf("!E get rows fail %+v", err)
		return err
	}
	isheader := true
	var headerIndex map[int]string
	indexs := make([]int, 0)
	for i := 1; i < len(rows); i++ {
		if isheader {
			headerIndex = getHeaderIndex(rows[i])
			for k := range headerIndex { //提取需要的index
				indexs = append(indexs, k)
			}
			isheader = false
			continue
		}
		keyTemp := make([]string, 2)
		var value types.BillItem
		for _, index := range indexs {
			switch headerIndex[index] {
			case constants.ProductType:
				keyTemp[0] = rows[i][index]
				value.ProductType = rows[i][index]
			case constants.ChargeType:
				keyTemp[1] = rows[i][index]
				value.ChargeType = rows[i][index]
			case constants.Cost:
				numT := strings.Split(rows[i][index][3:], ",")
				value.Cost, err = decimal.NewFromString(strings.Join(numT, "")) //去掉货币符号
				if err != nil {
					fmt.Println(rows[i][index][1:])
					log.Printf("!E parse fail %+v", err)
				}
			case constants.Arrears:
				numT := strings.Split(rows[i][index][3:], ",")
				value.Arrears, err = decimal.NewFromString(strings.Join(numT, ""))
				if err != nil {
					log.Printf("!E parse fail %+v", err)
				}
			}
		}
		if keyTemp[0] == "总计" {
			break
		}
		key := strings.Join(keyTemp, "-")
		if _, ok := p.LocalBills[key]; ok {
			p.LocalBills[key].Cost = p.LocalBills[key].Cost.Add(value.Cost)
			p.LocalBills[key].Arrears = p.LocalBills[key].Arrears.Add(value.Arrears)
		} else {
			p.LocalBills[key] = &value
		}
	}

	return nil
}

func (p *Processor) Compare() []types.Diff {
	result := make([]types.Diff, 0)
	for Ckey, Cvalue := range p.CloudBills {
		if Lvalue, ok := p.LocalBills[Ckey]; ok {
			if !Lvalue.Arrears.Equal(Cvalue.Arrears) || !Lvalue.Cost.Equal(Cvalue.Cost) || Lvalue.ChargeType != Cvalue.ChargeType || Lvalue.ProductType != Cvalue.ProductType {
				result = append(result, types.Diff{
					ProductDetail:       Cvalue.ProductType,
					ChargeType:          Cvalue.ChargeType,
					SchedulxCost:        Lvalue.Cost,
					AlibabaCloudCost:    Cvalue.Cost,
					SchedulxArrears:     Lvalue.Arrears,
					AlibabaCloudArrears: Cvalue.Arrears,
				})
			}
		} else { //cloudBill has but localBill doesn't
			result = append(result, types.Diff{
				ProductDetail:       Cvalue.ProductType,
				ChargeType:          Cvalue.ChargeType,
				SchedulxCost:        decimal.Zero,
				AlibabaCloudCost:    Cvalue.Cost,
				SchedulxArrears:     decimal.Zero,
				AlibabaCloudArrears: Cvalue.Arrears,
			})
		}
	}
	for Lkey, Lvalue := range p.LocalBills {
		if _, ok := p.CloudBills[Lkey]; !ok {
			result = append(result, types.Diff{
				ProductDetail:       Lvalue.ProductType,
				ChargeType:          Lvalue.ChargeType,
				SchedulxCost:        Lvalue.Cost,
				AlibabaCloudCost:    decimal.Zero,
				SchedulxArrears:     Lvalue.Arrears,
				AlibabaCloudArrears: decimal.Zero,
			})
		}
	}
	return result
}
func (p *Processor) ExportJSON(diffs []types.Diff) string {
	data := types.DataSet{diffs}
	result, err := json.Marshal(data)
	if err != nil {
		log.Printf("!E marshal fail %+v", err)
		return ""
	}
	return string(result)
}
func (p *Processor) ExportCSV(path string, diffs []types.Diff) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Printf("!E create file fail %+v", err)
		return
	}
	defer f.Close()
	writer := csv.NewWriter(f)
	header := []string{"产品明细", "付费类型", "SchedulX应付", "阿里云应付", "SchedulX欠费", "阿里云欠费"}
	writer.Write([]string{"\xEF\xBB\xBF"})
	err = writer.Write(header)
	writer.Flush()
	if err != nil {
		log.Printf("!E write csv header fail %+v", err)
		return
	}
	for _, v := range diffs {
		temp := []string{v.ProductDetail,
			v.ChargeType,
			v.SchedulxCost.String(),
			v.AlibabaCloudCost.String(),
			v.SchedulxArrears.String(),
			v.AlibabaCloudArrears.String()}
		err = writer.Write(temp)
		if err != nil {
			log.Printf("!E write one record fail %+v", err)
			continue
		}
		writer.Flush()
	}
	writer.Flush()

}
func getHeaderIndex(record []string) map[int]string {
	result := make(map[int]string, 4)
	for i, v := range record {
		if len(result) == 4 {
			return result
		}
		switch v {
		case "产品明细":
			result[i] = constants.ProductType
		case "付费类型", "消费类型":
			result[i] = constants.ChargeType
		case "应付金额":
			result[i] = constants.Cost
		case "欠费金额":
			result[i] = constants.Arrears
		}
	}
	return result
}
