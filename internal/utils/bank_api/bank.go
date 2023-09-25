package bank_api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"payment-go/internal/config"
	"payment-go/internal/transport/bank"
	"strings"
)

type IBank interface {
	CheckMerchant(proxy.IProxy, *bank.CheckMerchantRequestDto) (*bank.CheckMerchantResponseDto, error)
	GetAttributes(proxy.IProxy) (*bank.AttributesDto, error)
	GetLink(proxy.IProxy, *bank.LinkRequestDto) (*bank.LinkResponseDto, error)
}
type bankApi struct {
}

func NewBank() IBank {
	return &bankApi{}
}

func (b *bankApi) GetAttributes(pr proxy.IProxy) (*bank.AttributesDto, error) {
	res, err := pr.Get(config.BankLinks().GetAttsUrl)
	if err != nil {
		return nil, err
	}

	var atts bank.AttributesDto
	err = json.NewDecoder(res.Body).Decode(&atts)
	if err != nil {
		return nil, err
	}

	if err = atts.Validate(); err != nil {
		return nil, err
	}

	return &atts, nil
}

func (b *bankApi) CheckMerchant(pr proxy.IProxy, requestDto *bank.CheckMerchantRequestDto) (*bank.CheckMerchantResponseDto, error) {
	data, err := json.Marshal(requestDto)
	if err != nil {
		return nil, err
	}

	reader := strings.NewReader(string(data))
	res, err := pr.Post(config.BankLinks().CheckMerchantUrl, "application/json", reader)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var resp bank.CheckMerchantResponseDto
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	if err = resp.Validate(); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (b *bankApi) GetLink(pr proxy.IProxy, requestDto *bank.LinkRequestDto) (*bank.LinkResponseDto, error) {
	data, err := json.Marshal(requestDto)
	if err != nil {
		return nil, err
	}

	reader := strings.NewReader(string(data))
	res, err := pr.Post(config.BankLinks().PaymentUrl, "application/json", reader)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var resp bank.LinkResponseDto
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	if err = resp.Validate(); err != nil {
		return nil, err
	}

	return &resp, nil
}

func (b *bankApi) CheckLink(pr proxy.IProxy, transactionId string) (bool, error) {
	fmt.Println("transactionId", transactionId)
	fmt.Println("bank link transaction info", config.BankLinks().TransactionInfo+transactionId)
	res, err := pr.Get(config.BankLinks().TransactionInfo + transactionId)
	if err != nil {
		var data []byte
		res.Body.Read(data)
		fmt.Println(data)
		fmt.Println("CheckLink http error: ", err)
		return false, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Println(body)
		fmt.Println("CheckLink ioutil error: ", err)
		return false, err
	}

	var resp bank.CheckLinkResponseDto
	err = json.Unmarshal(body, &resp)
	if err != nil {
		fmt.Println("CheckLink json error: ", err)
		return false, err
	}

	fmt.Println(resp)

	return resp.Success(), nil
}
