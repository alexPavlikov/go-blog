package app

import (
	"crypto/rand"
	"fmt"

	"github.com/alexPavlikov/go-blog/models"
	"github.com/xuri/excelize/v2"
)

func Excel(sales []models.JoinStorePlus) {
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	// Create a new sheet.
	// index, err := f.NewSheet("Sheet2")
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// Set title of a cell.
	f.SetCellValue("Sheet1", "A1", "Id")
	f.SetCellValue("Sheet1", "B1", "Товар")
	f.SetCellValue("Sheet1", "C1", "Клиент")
	f.SetCellValue("Sheet1", "D1", "Адресс")
	f.SetCellValue("Sheet1", "E1", "Цена покупки")
	f.SetCellValue("Sheet1", "F1", "Категория товара")
	f.SetCellValue("Sheet1", "G1", "Пол")
	f.SetCellValue("Sheet1", "H1", "Дата")

	// alphbet := []string{"A", "B", "C", "D", "E", "F", "G", "H"}

	// Set value of a cell.
	for i, sls := range sales {
		f.SetCellValue("Sheet1", fmt.Sprintf("A%d", i+2), sls.Id)
		f.SetCellValue("Sheet1", fmt.Sprintf("B%d", i+2), fmt.Sprint(sls.Article)+"-"+sls.Name)
		f.SetCellValue("Sheet1", fmt.Sprintf("C%d", i+2), sls.User)
		f.SetCellValue("Sheet1", fmt.Sprintf("D%d", i+2), sls.Address)
		f.SetCellValue("Sheet1", fmt.Sprintf("E%d", i+2), sls.NewPrice)
		f.SetCellValue("Sheet1", fmt.Sprintf("F%d", i+2), sls.Category)
		f.SetCellValue("Sheet1", fmt.Sprintf("G%d", i+2), sls.Sex)
		f.SetCellValue("Sheet1", fmt.Sprintf("H%d", i+2), sls.Date)
	}

	// Set active sheet of the workbook.
	// f.SetActiveSheet(index)
	// Save spreadsheet by the given path.
	RandomCrypto, _ := rand.Prime(rand.Reader, 32)
	name := fmt.Sprint(RandomCrypto.Int64() / 2000)
	if err := f.SaveAs(fmt.Sprintf("C:/Users/admin/Desktop/%s.xlsx", name)); err != nil {
		fmt.Println(err)
	}
}
