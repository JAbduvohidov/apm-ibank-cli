package main

import (
	"database/sql"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/JAbduvohidov/apm-ibank-cli/cmd/common"
	"github.com/JAbduvohidov/apm-ibank-core/pkg/core"
	_ "github.com/mattn/go-sqlite3"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

const (
	jsonFormat    = ".json"
	xmlFormat     = ".xml"
	byName        = "byName"
	byPhoneNumber = "byPhoneNumber"
)

func main() {
	file, err := os.OpenFile("manager_log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
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

	fmt.Println( welcomeTitle)
	log.Println("start operations loop")
	operationsLoop(db, managersCommands)
	log.Println("finish operations loop")
	log.Println("finish application")
}

func operationsLoop(db *sql.DB, commands string) {
	for {
		fmt.Print( commands)
		cmd := common.GetCommand()
		log.Println("start of operation selection")
		common.ClearConsole()
		switch cmd {
		case "1":
			log.Println("add client operation selected")
			addClientToDb(db)
		case "2":
			log.Println("add account to client operation selected")
			addAccountToClient(db)
		case "3":
			log.Println("add service operation selected")
			addServiceToDb(db)
		case "4":
			log.Println("add atm operation selected")
			addAtmToDb(db)
		case "5":
			log.Println("export operation selected")
			exportOperationsLoop(db, exportImportCommands)
		case "6":
			log.Println("import operation selected")
			importOperations(db, exportImportCommands)
		case "7":
			log.Println("print list of clients by 10 operation selected")
			printListOfClients(db)
		case "8":
			log.Println("lock/unlock operation selected")
			changeClientStatus(db)
		case "9":
			log.Println("search client operation selected")
			searchClientOperations(db)
		case "q":
			log.Println("exit operation selected")
			return
		default:
			log.Println("incorrect operation selected")
			fmt.Printf("Вы выбрали неверную команду: %s\n", cmd)
		}
	}
}

func searchClientOperations(db *sql.DB) bool {
	fmt.Print( searchClientCommands)
	cmd := common.GetCommand()
	common.ClearConsole()
	switch cmd {
	case "1":
		log.Println("search client by name operation selected")
		searchClientBy(byName, db)
	case "2":
		log.Println("search client by phone number selected")
		searchClientBy(byPhoneNumber, db)
	case "q":
		log.Println("exit operation selected")
		return true
	default:
		log.Println("incorrect operation selected")
		fmt.Printf("Вы выбрали неверную команду: %s\n", cmd)
	}
	return false
}

func searchClientBy(searchType string, db *sql.DB) {
	var clients []core.Client
	var err error
	if searchType == byName {
		log.Println("asking to enter client name")
		fmt.Print( "Введите имя пользователя: ")
		name := common.GetStringInput()
		log.Println("client name entered")

		log.Println("trying to search clients by name")
		clients, err = core.SearchClientByName(name, db)
		if err != nil {
			log.Printf("unable to search client: %v", err)
			fmt.Println( "Поиск не удался")
			return
		}
	} else if searchType == byPhoneNumber {
		log.Println("asking to enter clients' phone number")
		fmt.Print( "Введите номер телефона пользователя: ")
		phoneNumber := common.GetIntegerInput()
		log.Println("clients' phone number entered")

		log.Println("trying to search clients by phone number ")
		clients, err = core.SearchClientByPhoneNumber(phoneNumber, db)
		if err != nil {
			log.Printf("unable to search client: %v", err)
			fmt.Println( "Поиск не удался")
			return
		}
	}
	log.Println("search completed")
	if clients == nil {
		log.Println("nothing was found")
		fmt.Println( "Ничего не найдено")
		return
	}

	for indx, client := range clients {
		fmt.Println( indx+1, ") ", client.Name, client.PhoneNumber, client.Status)
	}
}

func changeClientStatus(db *sql.DB) {
	log.Println("asking to enter client phone number")
	fmt.Print( "Введите номер телефона пользователя: ")
	phoneNumber := common.GetIntegerInput()
	log.Println("clients' phone number entered")

	log.Println("asking to set status to client")
	fmt.Print( "Выберите статус пользователю (locked/active): ")
	status := common.GetCommand()
	status = strings.ToLower(status)

	if status == core.Active || status == core.Locked {
		log.Println("status set")
		log.Println("start changing client status")
		err := core.ChangeClientStatus(phoneNumber, status, db)
		if err != nil {
			if errors.Is(err, core.ErrPhoneNumberNotExist) {
				log.Println("phone number does not exist")
				fmt.Println( "Номер телефона не существует!")
			}
			log.Println("unable to change status")
			fmt.Println( "Не удалось изменить статус.")
			return
		}
		fmt.Println( "Статус изменён")
		return
	}

	log.Println("invalid status")
	fmt.Println( "Неверный статус")
}

func printListOfClients(db *sql.DB) {
	log.Println("start paging")
	var page int64
	page = 1
	for {
		offset := page * 10
		if page == 1 {
			offset = 0
		}
		log.Println("start getting list of clients")
		clients, err := core.GetListOfClientsFormatted(10, offset, db)
		if err != nil {
			log.Printf("unable to get list of clients: %v", err)
			fmt.Println( "Не удалось получить список пользователей!")
			return
		}
		log.Println("list of clients received")
		if clients == nil {
			log.Println("list of clients is empty")
			fmt.Println( "Пусто")
			return
		}

		for indx, client := range clients {
			fmt.Println( indx+1, ") ", client.Name, client.Login, client.PhoneNumber, client.Status)
		}

		fmt.Println( )
		fmt.Print( pagingOperations)
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

func importOperations(db *sql.DB, commands string) {
	fmt.Println( importTitle)
	fmt.Print( commands)
	cmd := common.GetCommand()
	common.ClearConsole()
	switch cmd {
	case "1":
		log.Println("import list of clients selected")
		var clients []core.Client
		unmarshaler(clients)

		log.Println("start importing list of clients to db")
		err := core.ImportListOfClients(clients, db)
		if err != nil {
			log.Printf("unable to import list of clients: %v", err)
			return
		}
		log.Println("list of clients imported to db")
		fmt.Println( "Список пользователей импортирован!")
	case "2":
		log.Println("import list of accounts with client ids selected")
		var accountWithClientIds []core.AccountWithClientId
		unmarshaler(accountWithClientIds)

		log.Println("start importing list of accountWithClientIds to db")
		err := core.ImportListOfAccounts(accountWithClientIds, db)
		if err != nil {
			log.Printf("unable to import list of accountWithClientIds: %v", err)
			return
		}
		log.Println("list of accountWithClientIds imported to db")
		fmt.Println( "Список аккаунтов с пользователями импортирован!")
	case "3":
		log.Println("import list of ATMs selected")
		var atms []core.ATM
		unmarshaler(atms)

		log.Println("start importing list of atms to db")
		err := core.ImportListOfATMs(atms, db)
		if err != nil {
			log.Printf("unable to import list of atms: %v", err)
			return
		}
		log.Println("list of atms imported to db")
		fmt.Println( "Список банкоматов импортирован!")
	case "q":
		log.Println("exit operation selected")
		return
	default:
		log.Println("incorrect operation selected")
		fmt.Printf("Вы выбрали неверную команду: %s\n", cmd)
	}
}

func unmarshaler(importTo interface{}) {
	fmt.Println( importTitle)
	log.Println("asking for full file path")
	fmt.Print( "Введите полный путь к файлу: ")
	fullPath := common.GetStringInput()
	log.Println("full file path entered")

	log.Println("start reading from file")
	file, err := ioutil.ReadFile(fullPath)
	if err != nil {
		log.Printf("can't read from file: %v", err)
		return
	}
	log.Println("end of reading from file")
	log.Println("detecting file format")
	if strings.HasSuffix(fullPath, jsonFormat) {
		log.Println("start unmarshal")
		err = json.Unmarshal(file, &importTo)

	} else if strings.HasSuffix(fullPath, xmlFormat) {
		log.Println("start unmarshal")
		err = xml.Unmarshal(file, &importTo)
	} else {
		log.Println("invalid file format")
		fmt.Println( "Неверный формат файла!")
		return
	}

	if err != nil {
		log.Printf("can't unmarshal: %v", err)
		return
	}
	log.Println("end of unmarshal")
	common.ClearConsole()
}

func exportOperationsLoop(db *sql.DB, commands string) {
	for {
		fmt.Println( exportTitle)
		fmt.Print( commands)
		cmd := common.GetCommand()
		common.ClearConsole()
		switch cmd {
		case "1":
			log.Println("export list of clients selected")

			log.Println("start getting list of clients")
			clients, err := core.GetListOfClients(db)
			if err != nil {
				log.Printf("unable to get list of clients: %v", err)
				fmt.Println( "Не удалось получить список пользователей")
				return
			}
			log.Println("list of clients received")
			if clients == nil {
				log.Println("list of clients is empty. No need for export.")
				fmt.Println( "Список пользователей пуст. Нечего экспортировать!")
				return
			}
			fileFormatOperations(fileFormats, core.Clients, clients)
		case "2":
			log.Println("export list of accounts with client ids selected")

			log.Println("start getting list of accounts with client ids")
			accountsWithClientIds, err := core.GetListOfAccountsWithClients(db)
			if err != nil {
				log.Printf("unable to get list of accounts with client ids: %v", err)
				fmt.Println( "Не удалось получить список аккаунтов с пользователями")
				return
			}
			log.Println("list of accounts with client ids received")
			if accountsWithClientIds == nil {
				log.Println("list of accounts with client ids is empty. No need for export.")
				fmt.Println( "Список аккаунтов с пользователями пуст. Нечего экспортировать!")
				return
			}
			fileFormatOperations(fileFormats, core.Accounts, accountsWithClientIds)
		case "3":
			log.Println("export list of ATMs selected")

			log.Println("start getting list of ATMs")
			listOfATMs, err := core.GetListOfATMs(db)
			if err != nil {
				log.Printf("unable to get list of ATMs: %v", err)
				fmt.Println( "Не удалось получить список банкоматов")
				return
			}
			log.Println("list of ATMs received")
			if listOfATMs == nil {
				log.Println("list of ATMs is empty. No need for export.")
				fmt.Println( "Список банкоматов пуст. Нечего экспортировать!")
				return
			}
			fileFormatOperations(fileFormats, core.ATMs, listOfATMs)
		case "q":
			log.Println("exit operation selected")
			return
		default:
			log.Println("incorrect operation selected")
			fmt.Printf("Вы выбрали неверную команду: %s\n", cmd)
		}
	}
}

func fileFormatOperations(formats string, title string, toExport interface{}) {
	fmt.Println( formatsTitle)
	fmt.Print( formats)
	cmd := common.GetCommand()
	common.ClearConsole()
	switch cmd {
	case "1":
		exportTo(jsonFormat, title, toExport)
	case "2":
		exportTo(xmlFormat, title, toExport)
	case "q":
		log.Println("exit operation selected")
		return
	default:
		log.Println("incorrect operation selected")
		fmt.Printf("Вы выбрали неверную команду: %s\n", cmd)
	}
}

func exportTo(format string, title string, export interface{}) {
	log.Println("export started")
	fileName := title + format
	var err error = nil
	var marshal []byte
	switch format {
	case jsonFormat:
		marshal, err = json.Marshal(export)
	case xmlFormat:
		marshal, err = xml.Marshal(export)
	default:
		fmt.Println( "Неверный формат.")
		return
	}
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("exporting clients to \"%s\" format", format)
	err = ioutil.WriteFile(fileName, marshal, 0666)
	log.Println("file exported")
	fmt.Println( "Файл экспортирован.")
}

func addAtmToDb(db *sql.DB) {
	fmt.Println( addingAtmTitle)
	log.Println("asking to enter byName of ATM")
	fmt.Print( "Введите название банкомата: ")
	nameOfAtm := common.GetStringInput()
	log.Println("byName of ATM entered")

	log.Println("asking to enter location of ATM")
	fmt.Print( "Введите расположение банкомата: ")
	locationOfAtm := common.GetStringInput()
	log.Println("location of ATM entered")

	log.Println("start adding ATM to db")
	err := core.AddAtm(nameOfAtm, locationOfAtm, db)
	if err != nil {
		log.Printf("unable to add ATM to db: %v", err)
		common.ClearConsole()
		fmt.Println( "Не удалось добавить банкомат.")
		if errors.Is(err, core.ErrATMExist) {
			fmt.Printf("Банкомат в \"%s\"-е уже еcть\n", locationOfAtm)
		}
		return
	}
	common.ClearConsole()
	log.Println("ATM added to db")
	fmt.Printf("Банкомат %s добавлен!\n", nameOfAtm)
}

func addServiceToDb(db *sql.DB) {
	fmt.Println( addingServiceTitle)
	log.Println("asking to enter byName of service")
	fmt.Print( "Введите название услуги: ")
	nameOfService := common.GetStringInput()
	log.Println("byName of service entered")
	log.Println("start adding service to db")
	err := core.AddService(nameOfService, db)
	if err != nil {
		log.Printf("unable to add service to db: %v", err)
		common.ClearConsole()
		fmt.Println( "Не удалось добавить новую услугу")
		if errors.Is(err, core.ErrServiceExist) {
			fmt.Printf("Услуга \"%s\" существует\n", nameOfService)
		}
		return
	}
	common.ClearConsole()
	log.Println("service added to db")
	fmt.Printf("Услуга \"%s\" добавлена!\n", nameOfService)
}

func addAccountToClient(db *sql.DB) {
	fmt.Println( addingAccountToClientTitle)
	log.Println("asking to enter phone number")
	fmt.Print( "Введите номер телефона: ")
	phoneNumber := common.GetIntegerInput()
	log.Println("phone number entered")

	log.Println("asking to enter cash amount to add to account")
	fmt.Print( "Введите сумму в рублях: ")
	balance := common.GetIntegerInput()
	log.Println("cash amount entered")

	log.Println("start adding account to client")
	err := core.AddAccount(phoneNumber, balance, db)
	if err != nil {
		log.Printf("unable to add account to client: %v", err)
		common.ClearConsole()
		fmt.Printf("Не удалось добавить счёт на номер \"%d\"\n", phoneNumber)
		return
	}
	common.ClearConsole()
	log.Println("account added")
	fmt.Printf("Добавлен счёт на номер \"%d\"\n", phoneNumber)
}

func addClientToDb(db *sql.DB) {
	fmt.Println( addingClientTitle)
	fmt.Print( "Введите имя: ")
	log.Println("asking for byName")
	name := common.GetStringInput()
	log.Println("byName entered")

	fmt.Print( "Введите номер телефона: ")
	log.Println("asking for phone number")
	phoneNumber := common.GetIntegerInput()
	log.Println("phone number entered")

	fmt.Print( "Придумайте логин: ")
	log.Println("asking to create login")
	login := common.GetStringInput()
	log.Println("login created")

	fmt.Print( "Придумайте надёжный пароль: ")
	log.Println("asking to create password")
	password := common.GetStringInput()
	log.Println("password entered")

	log.Println("adding client to db")



	err := core.AddClient(name, login, password, phoneNumber, db)
	if err != nil {
		log.Printf("unable to add client: %v", err)
		common.ClearConsole()
		fmt.Println( "Не удалось добавить нового пользователя")
		if errors.Is(err, core.ErrLoginExist) {
			fmt.Println( "Пользователь с таким логином существует.")
		}
		if errors.Is(err, core.ErrPhoneNumberExist) {
			fmt.Println( "Пользователь с таким номером существует")
		}
		return
	}
	common.ClearConsole()
	log.Println("client added to db")
	fmt.Printf("Пользователь \"%s\" добавлен!\n", name)
}