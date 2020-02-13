package main

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/JAbduvohidov/apm-ibank-cli/cmd/common"
	"github.com/JAbduvohidov/apm-ibank-core/pkg/core"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
)

func main() {
	file, err := os.OpenFile("client_log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	log.Print("start application")
	log.Print("start opening db")
	db, err := sql.Open("sqlite3", "db.sqlite")
	if err != nil {
		log.Fatalf("can't open db: %v", err)
	}
	log.Print("db opened")
	defer func() {
		log.Print("start closing db")
		if err := db.Close(); err != nil {
			log.Fatalf("can't close db: %v", err)
		}
		log.Println("db closed")

		if err = file.Close(); err != nil {
			log.Fatalf("error closing file: %v", err)
		}
	}()
	log.Println("initialising db")
	err = core.Init(db)
	if err != nil {
		log.Fatalf("can't initialise db: %v", err)
	}
	log.Println("db initialised")

	fmt.Println(welcomeTitle)
	log.Println("unauthorised operations loop started")
	unauthorisedOperationsLoop(db, unauthorisedOperations)
	log.Println("finish unauthorised operations loop")
	log.Println("finish application")
}

func unauthorisedOperationsLoop(db *sql.DB, commands string) {
	for {
		fmt.Print(commands)
		cmd := common.GetCommand()
		common.ClearConsole()
		switch cmd {
		case "1":
			log.Println("login operation selected")
			loginOperations(db)
		case "2":
			log.Println("get list of atms operation selected")
			printListOfATMs(db)
		case "q":
			log.Println("exit operation selected")
			return
		default:
			log.Println("incorrect operation selected")
			fmt.Printf("Вы выбрали неверную команду: %s\n", cmd)
		}
	}
}

func printListOfATMs(db *sql.DB) {
	log.Println("start getting list of atms")
	listOfATMs, err := core.GetListOfATMs(db)
	if err != nil {
		log.Printf("unable to get list of atms: %v", err)
		fmt.Println("Не удалось получить список бакоматов!")
		return
	}
	log.Println("list of atms received")

	if listOfATMs == nil {
		log.Print("list of atms is empty")
		fmt.Println("Список банкоматов пуст.")
		return
	}
	for idx, atm := range listOfATMs {
		fmt.Printf("%d) Название: %s расположение: %s\n", idx+1, atm.Name, atm.Location)
	}
}

func printListOfAccounts(login string, db *sql.DB) {
	accounts, err := core.GetListOfClientAccounts(login, db)
	if err != nil {
		log.Printf("unable to get list of client accounts")
		fmt.Println("Не удалось получить список счетов")
		return
	}
	if accounts == nil {
		log.Println("list of accounts is empty")
		fmt.Println("Список счетов пуст")
		return
	}
	for idx, account := range accounts {
		fmt.Printf("%d) Счет: %d баланc: %f\n", idx+1, account.Id, account.Balance)
	}
}

func loginOperations(db *sql.DB) {
	fmt.Println(loginTitle)
	log.Print("asking to enter login")
	fmt.Print("Введите логин: ")
	login := common.GetStringInput()
	log.Print("login entered")

	log.Print("asking to enter password")
	fmt.Print("Введите пароль: ")
	password := common.GetStringInput()
	log.Print("password entered")

	log.Print("trying to login")
	phoneNumber, err := core.Login(login, password, db)
	if err != nil {
		if errors.Is(err, core.ErrInvalidPass) {
			log.Println("invalid password")
			fmt.Println("Неверный пароль.")
		}
		if errors.Is(err, core.ErrClientIsLocked) {
			log.Println("client is locked")
			fmt.Println("Просим прощения, но ваш аккаунт был заблокирован по каким-то серьёзным причинам (")
		}
		log.Printf("unable to login: %v", err)
		return
	}
	if phoneNumber == -1 {
		log.Print("invalid login or password")
		fmt.Println("Неверный логин или пароль.")
		return
	}
	log.Print("login success")
	log.Print("authorised operations loop started")
	authorisedOperationsLoop(phoneNumber, login, authorisedOperations, db)
	log.Print("authorised operations loop ended")
}

func authorisedOperationsLoop(phoneNumber int64, login, commands string, db *sql.DB) {
	for {
		fmt.Print(commands)
		cmd := common.GetCommand()
		common.ClearConsole()
		switch cmd {
		case "1":
			log.Println("get list of client accounts")
			printListOfAccounts(login, db)
		case "2":
			log.Println("transfer money operation selected")
			transferMoneyOperationsLoop(login, phoneNumber, db)
		case "3":
			log.Println("pay for service operation selected")
			payForService(login, db)
		case "4":
			log.Println("get journal list operation selected")
			printJournalListOperationsLoop(login, db)
		case "q":
			log.Println("exit operation selected")
			return
		default:
			log.Println("incorrect operation selected")
			fmt.Printf("Вы выбрали неверную команду: %s\n", cmd)
		}
	}
}

func printJournalListOperationsLoop(login string, db *sql.DB) {
	log.Println("start paging")
	var page int64
	page = 1
	for {
		offset := page * 10
		if page == 1 {
			offset = 0
		}
		log.Println("start getting list of journals")
		journals, err := core.GetJournalListFormatted(login, 10, offset, db)
		if err != nil {
			log.Printf("unable to get list of journals: %v", err)
			fmt.Println("Не удалось получить журнал операций!")
			return
		}
		log.Println("list of journals received")
		if journals == nil {
			log.Println("list of journals is empty")
			fmt.Println("Пусто")
			return
		}

		for indx, journal := range journals {
			fmt.Println(indx+1, ") ", journal.Date, journal.Type, journal.TransferredTo, journal.Amount)
		}

		fmt.Println()
		fmt.Print(pagingOperations)
		cmd := common.GetCommand()
		common.ClearConsole()
		switch cmd {
		case "1":
			log.Println("next operation selected")
			page++
		case "2":
			log.Println("prev operation selected")
			if page != 0 {
				page--
			}
		case "q":
			log.Println("exit operation selected")
			return
		default:
			log.Println("incorrect operation selected")
			fmt.Printf("Вы выбрали неверную команду: %s\n", cmd)
		}
	}
}

func transferMoneyOperationsLoop(login string, phoneNumber int64, db *sql.DB) {
	fmt.Println(transferTitle)
	fmt.Print(transferMoneyOperations)
	cmd := common.GetCommand()
	common.ClearConsole()
	switch cmd {
	case "1":
		log.Println("transfer money by account id operation selected")

		transferByAccount(login, db)
	case "2":
		log.Println("transfer money by phone number operation selected")

		transferByPhoneNumber(phoneNumber, login, db)
	case "q":
		log.Println("exit operation selected")
		return
	default:
		log.Println("incorrect operation selected")
		fmt.Printf("Вы выбрали неверную команду: %s\n", cmd)
	}
}

func transferByPhoneNumber(phoneNumber int64, login string, db *sql.DB) {
	fmt.Println(transferTitle)
	log.Println("asking to enter target phone number")
	fmt.Print("Введите номер телефона цели: ")
	targetPhoneNumber := common.GetIntegerInput()
	log.Println("target phone number entered")

	log.Println("asking to enter account id")
	fmt.Print("Введите номер счета: ")
	accountId := common.GetIntegerInput()
	log.Println("account id entered")

	log.Println("asking to enter amount")
	fmt.Print("Введите сумму для перевода: ")
	amount := common.GetIntegerInput()
	log.Println("amount entered")

	log.Println("trying to transfer money")
	if phoneNumber == targetPhoneNumber {
		log.Println("can't transfer money to the same person")
		fmt.Println("Перевод средств на свой номер невозможен!")
		return
	}

	ok, err := checkAccountIfValid(login, db, accountId)
	if !ok {
		log.Printf("unable to transfer money: %v", err)
		fmt.Println("Перевод средств не удался!")
		return
	}

	err = core.TransferToByPhoneNumber(targetPhoneNumber, login, accountId, float64(amount), db)
	if err != nil {
		if errors.Is(err, core.ErrClientIsLocked) {
			log.Println("target client is locked")
			fmt.Println("Пользователь заблокирован!")
		}
		log.Printf("can't transfer money: %v", err)
		fmt.Println("Перевод средств невозможен")
		return
	}
	log.Println("money transferred")
	fmt.Println("Средства начислены!")
}

func checkAccountIfValid(login string, db *sql.DB, accountId int64) (ok bool, err error) {
	ok = false
	accounts, err := core.GetListOfClientAccounts(login, db)
	for _, account := range accounts {
		if account.Id == accountId {
			ok = true
		}
	}
	return ok, err
}

func transferByAccount(login string, db *sql.DB) {
	fmt.Println(transferTitle)
	log.Print("asking to enter account id")
	fmt.Print("Введите свой номер счета: ")
	accountId := common.GetIntegerInput()
	log.Print("account id entered")

	log.Print("asking to enter amount")
	fmt.Print("Введите сумму для перевода средств: ")
	amount := common.GetIntegerInput()
	log.Print("amount entered")

	log.Print("asking to enter target account id")
	fmt.Print("Введите номер счета цели: ")
	targetAccountId := common.GetIntegerInput()
	log.Print("target account id entered")

	common.ClearConsole()

	log.Print("trying to transfer money")
	if accountId == targetAccountId {
		log.Print("can't transfer money to the same account")
		fmt.Println("Вы не можете перевести средства на один и тот-же счет.")
		return
	}

	ok, err := checkAccountIfValid(login, db, accountId)

	if !ok {
		log.Printf("unable to transfer money: %v", err)
		fmt.Println("Перевод средств не удался!")
		return
	}

	err = core.TransferToByAccountId(targetAccountId, login, accountId, float64(amount), db)
	if err != nil {
		if errors.Is(err, core.ErrClientIsLocked) {
			log.Println("target client is locked")
			fmt.Println("Пользователь заблокирован.")
		}
		log.Printf("unable to transfer money: %v", err)
		fmt.Println("Перевод средств не удался!")
		return
	}
	log.Print("money transferred")
	fmt.Println("Средства начислены.")
}

func payForService(login string, db *sql.DB) {
	fmt.Println(payForServiceTitle)
	log.Println("asking to enter account id")
	fmt.Print("Введите номер счета: ")
	accountId := common.GetIntegerInput()
	log.Println("account id entered")
	log.Println("asking to enter payment amount")
	fmt.Print("Введите оплачеваемую сумму: ")
	amount := common.GetIntegerInput()
	log.Println("payment amount entered")
	log.Println("asking to enter name of service")
	fmt.Print("Введите название услуги: ")
	nameOfService := common.GetStringInput()
	log.Println("name of service entered")
	log.Println("trying to pay for service")
	err := core.PayForService(nameOfService, accountId, login, float64(amount), db)
	common.ClearConsole()
	if err != nil {
		if errors.Is(err, core.ErrServiceNotExist) {
			log.Println("service does not exist")
			fmt.Println("Данная услуга не существует.")
			return
		}
		log.Printf("unable to pay for service: %v", err)
		fmt.Println("Не удалось оплатить услугу.")
		return
	}
	log.Println("payment done")
	fmt.Printf("Услуга \"%s\" оплачена!\n", nameOfService)
}
