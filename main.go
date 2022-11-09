package main

import (
	"flag"
	"fmt"
	"github.com/galayx-future/billcompare/internal/service"
	"github.com/galayx-future/billcompare/tools"
)

var (
	pathCloudBill string
	pathLocalBill string
	pathExport    string
	needCSV       bool
)

func main() {
	flag.StringVar(&pathCloudBill, "cloud", "", "云厂商账单(.csv)所在路径")
	flag.StringVar(&pathLocalBill, "local", "", "schedulx账单(.xlsx)所在路径")
	flag.StringVar(&pathExport, "export", "./compare_result.csv", "将输出文件(.csv)导出位置")
	flag.BoolVar(&needCSV, "csv", false, "是否需要导出csv文件")
	flag.Parse()
	p := service.NewProcessor()
	if err := p.XLSXLoadLocal(pathLocalBill); err != nil {
		return
	}
	if err := p.CSVLoadCloud(pathCloudBill); err != nil {
		return
	}
	diff := p.Compare()
	diff = tools.CleanUp(diff)
	if len(diff) == 0 {
		fmt.Println("两账单一致")
	}
	if needCSV {
		p.ExportCSV(pathExport, diff)
	}
	fmt.Println(p.ExportJSON(diff))
}
