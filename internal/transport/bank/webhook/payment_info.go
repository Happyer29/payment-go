package webhook

import (
	"encoding/json"
	"errors"
	"fmt"
	"payment-go/internal/config"
	"payment-go/internal/models"
	"regexp"
	"strconv"
	"strings"
)

type BankPaymentInfoRequest struct {
	From       *string `json:"from,omitempty"`
	To         *string `json:"to,omitempty"`
	Content    string  `json:"content,omitempty"`
	Dir        *string `json:"dir,omitempty"`
	Password   string  `json:"password,omitempty"`
	CardNumber string  `json:"card_number,omitempty"`
	Amount     string  `json:"amount,omitempty"`
}

var ErrInvalidJson = errors.New("invalid json")
var ErrEmptyPassword = errors.New("password cannot be empty")

func FromRequestBody(data []byte) (*BankPaymentInfoRequest, error) {
	var request *BankPaymentInfoRequest
	if err := json.Unmarshal(data, &request); err != nil {
		return nil, ErrInvalidJson
	}
	return request, nil
}

func FromPostFields(provider func(field string, defaultValue ...string) string) (*BankPaymentInfoRequest, error) {
	from := provider("from")
	to := provider("to")
	dir := provider("dir")
	req := &BankPaymentInfoRequest{
		From:       &from,
		To:         &to,
		Content:    provider("content"),
		Dir:        &dir,
		Password:   provider("password"),
		CardNumber: provider("card_number"),
		Amount:     provider("amount"),
	}

	if len(req.Password) == 0 {
		return nil, ErrEmptyPassword
	}

	return req, nil
}

func (req *BankPaymentInfoRequest) isTypeRaw() bool {
	return len(req.Password) == 0 && len(req.CardNumber) == 0 && len(req.Amount) == 0
}

func (req *BankPaymentInfoRequest) GetRawContent() string {
	return req.Content
}
func (req *BankPaymentInfoRequest) ToDto() (*BankPaymentInfoDto, error) {
	return FromRequest(req)
}

type BankPaymentInfoDto struct {
	From       *string
	To         *string
	RawContent string
	Password   string
	CardNumber string
	AmountRaw  string
	Amount     float64
	Message    string
}

func FromRequest(request *BankPaymentInfoRequest) (*BankPaymentInfoDto, error) {
	if request.isTypeRaw() {
		return fromRawRequest(request)
	} else {
		return fromJsonRequest(request)
	}
}

func fromRawRequest(request *BankPaymentInfoRequest) (*BankPaymentInfoDto, error) {
	content := request.Content

	dto := &BankPaymentInfoDto{
		From:       request.From,
		To:         request.To,
		RawContent: content,
	}

	var lines []string
	for _, line := range strings.Split(content, "\n") {
		line = strings.Trim(line, " ")
		if len(line) > 0 {
			lines = append(lines, line)
		}
	}

	if len(lines) < 3 {
		return nil, fmt.Errorf("too short message")
	}

	dto.Password = parsePassword(lines[0])
	dto.CardNumber = parseCardNumber(lines[1])
	dto.AmountRaw = lines[2]
	if len(lines) > 3 {
		dto.Message = strings.Join(lines[3:], "\n")
	}

	amount, err := parseAmount(dto.AmountRaw)
	if err != nil {
		return nil, err
	}
	dto.Amount = amount

	if err = dto.Validate(); err != nil {
		return nil, err
	}

	return dto, nil
}

func fromJsonRequest(request *BankPaymentInfoRequest) (*BankPaymentInfoDto, error) {
	amount, err := parseAmount(request.Amount)
	if err != nil {
		return nil, err
	}

	dto := &BankPaymentInfoDto{
		From:       request.From,
		To:         request.To,
		RawContent: request.Content,
		Password:   parsePassword(request.Password),
		CardNumber: parseCardNumber(request.CardNumber),
		AmountRaw:  request.Amount,
		Amount:     amount,
		Message:    request.Content,
	}
	return dto, nil
}

func parseAmount(amountRaw string) (float64, error) {
	amountStr := regexp.MustCompile("[^0-9.,]+").ReplaceAllString(amountRaw, "")
	amount, err := strconv.ParseFloat(amountStr, 32)
	if err != nil {
		return 0, fmt.Errorf("unable to parse amount")
	}
	return amount, nil
}

func parseCardNumber(raw string) string {
	regexCardNumber := regexp.MustCompile("[^0-9]+")
	return regexCardNumber.ReplaceAllString(raw, "")
}

func parsePassword(raw string) string {
	return strings.Trim(raw, " ")
}

func (dto *BankPaymentInfoDto) Validate() error {
	if dto.Amount <= 0 {
		return fmt.Errorf("amount must be positive float")
	}
	if len(dto.CardNumber) != 16 {
		return fmt.Errorf("card number must be 16 digits")
	}
	if len(dto.Password) == 0 {
		return fmt.Errorf("password cannot be empty")
	}
	if dto.Password != config.BankSMSWebhookPassword {
		return fmt.Errorf("invalid password")
	}
	return nil
}

func (dto *BankPaymentInfoDto) ToBankMessage() *models.BankMessage {
	return &models.BankMessage{
		SenderPhone:   dto.From,
		ReceiverPhone: dto.To,
		RawMessage:    dto.RawContent,
		CardNumber:    dto.CardNumber,
		Amount:        dto.Amount,
	}
}
