package services

import (
	"github.com/menklab/goCMS/repositories"
	"log"
	"github.com/menklab/goCMS/models"
	"time"
	"github.com/menklab/goCMS/config"
	"fmt"
	"github.com/menklab/goCMS/utility/errors"
)

type IEmailService interface {
	SetVerified(email string) error
	GetVerified(email string) bool
	AddEmail(*models.Email) error
	SendEmailActivationCode(email string) error
	VerifyEmailActivationCode(id int, code string) bool
	// Promote Email
}

type EmailService struct {
	MailService IMailService
	AuthService IAuthService
	RepositoriesGroup *repositories.RepositoriesGroup
}


func DefaultEmailService(rg *repositories.RepositoriesGroup, ms *MailService, as *AuthService) *EmailService {
	emailService := &EmailService{
		RepositoriesGroup: rg,
		AuthService: as,
		MailService: ms,
	}
	return emailService
}

func (es *EmailService) SetVerified (e string) error {
	// get email
	email, err := es.RepositoriesGroup.EmailRepository.GetByAddress(e)
	if err != nil {
		return err
	}

	// set verified
	email.IsVerified = true
	err = es.RepositoriesGroup.EmailRepository.Update(email)
	if err != nil {
		return err
	}
	return err
}

func (es *EmailService) GetVerified (e string) bool {
	email, err := es.RepositoriesGroup.EmailRepository.GetByAddress(e)
	if err != nil {
		log.Printf("Email Service, Get Verified, Error getting email by address: %s\n", err.Error())
		return false
	}

	return email.IsVerified
}

func (es *EmailService) AddEmail (e *models.Email) error {

	// check to see if email exist
	emailExists, _ := es.RepositoriesGroup.EmailRepository.GetByAddress(e.Email)
	if emailExists != nil {
		return errors.NewToUser("Email already exists.")
	}

	// add email
	err := es.RepositoriesGroup.EmailRepository.Add(e)
	if err != nil {
		log.Printf("Email Service, error adding email: %s", err.Error())
		return err
	}

	return nil
}

func (es *EmailService) SendEmailActivationCode(emailAddress string) error {

	// get userId from email
	email, err := es.RepositoriesGroup.EmailRepository.GetByAddress(emailAddress)
	if err != nil {
		fmt.Printf("Error sending email activation code, get email: %s", err.Error())
		return err
	}

	if email.IsVerified {
		err = errors.NewToUser("Email already activated.")
		fmt.Printf("Error sending email activation code, %s\n", err.Error())
		return err
	}

	// create reset code
	code, hashedCode, err := es.AuthService.GetRandomCode(32)
	if err != nil {
		fmt.Printf("Error sending email activation code, get random code: %s\n", err.Error())
		return err
	}

	// update user with new code
	err = es.RepositoriesGroup.SecureCodeRepository.Add(&models.SecureCode{
		UserId: email.UserId,
		Type:   models.Code_VerifyEmail,
		Code:   hashedCode,
	})
	if err != nil {
		fmt.Printf("Error sending email activation code, add secure code: %s\n", err.Error())
		return err
	}

	// send email
	es.MailService.Send(&Mail{
		To:      emailAddress,
		Subject: "Email Verification Required",
		Body: "Click on the link below to activate your email:\n" +
			config.PublicApiUrl + "/activate-email?code=" + code + "&email=" + emailAddress + "\n\nThe link will expire at: " +
			time.Now().Add(time.Minute * time.Duration(config.PasswordResetTimeout)).String() + ".",
	})
	if err != nil {
		log.Println("Error sending email activation code, sending mail: " + err.Error())
	}

	return nil
}

func (es *EmailService) VerifyEmailActivationCode(id int, code string) bool {

	// get code
	secureCode, err := es.RepositoriesGroup.SecureCodeRepository.GetLatestForUserByType(id, models.Code_VerifyEmail)
	if err != nil {
		log.Printf("error getting latest password reset code: %s", err.Error())
		return false
	}

	if ok := es.AuthService.VerifyPassword(secureCode.Code, code); !ok {
		return false
	}

	// check within time
	if time.Since(secureCode.Created) > (time.Minute * time.Duration(config.PasswordResetTimeout)) {
		return false
	}

	err = es.RepositoriesGroup.SecureCodeRepository.Delete(secureCode.Id)
	if err != nil {
		return false
	}

	return true
}