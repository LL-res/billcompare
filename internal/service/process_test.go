package service

import (
	"fmt"
	"github.com/galayx-future/billcompare/tools"
	"testing"
)

func TestProcessor_XLSXLoadLocal(t *testing.T) {
	p := NewProcessor()
	if err := p.XLSXLoadLocal("C:\\Users\\LL\\Desktop\\local.xlsx"); err != nil {
		t.Error(err)
		return
	}
	for k, v := range p.LocalBills {
		fmt.Printf("%v: %v\n", k, v)
	}
}
func TestProcessor_CSVLoadCloud(t *testing.T) {
	p := NewProcessor()
	if err := p.CSVLoadCloud("C:\\Users\\LL\\Desktop\\cloudx.csv"); err != nil {
		t.Error(err)
		return
	}
	for k, v := range p.CloudBills {
		fmt.Printf("%v: %v\n", k, v)
	}
}
func TestProcessor_Compare(t *testing.T) {
	p := NewProcessor()
	if err := p.XLSXLoadLocal("C:\\Users\\LL\\Desktop\\local.xlsx"); err != nil {
		t.Error(err)
		return
	}
	if err := p.CSVLoadCloud("C:\\Users\\LL\\Desktop\\cloudx.csv"); err != nil {
		t.Error(err)
		return
	}
	diff := p.Compare()
	p.ExportCSV("./compare_result.csv", tools.CleanUp(diff))
}
