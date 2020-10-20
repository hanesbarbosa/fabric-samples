/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/hanesbarbosa/phe"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SimpleContract provides functions for managing a car
type SimpleContract struct {
	contractapi.Contract
}

// Patient describes basic details of a patient
type Patient struct {
	Name                  string `json:"name"`
	PreExistingConditions string `json:"preExistingConditions"`
	DiagnosisID           string `json:"diagnosisID"`
	StatusID              string `json:"statusID"`
	KeyID                 string `json:"keyID"`
}

// Proposal ...
type Proposal struct {
	RequesterID string `json:"requesterID"`
	RequestedID string `json:"requestedID"`
	PatientsIDs string `json:"patientsIDs"`
	KeyID       string `json:"keyID"`
	Value       string `json:"value"`
}

// Result ...
type Result struct {
	ProposalID string `json:"proposalID"`
	KeyID      string `json:"keyID"`
	Value      string `json:"value"`
}

// QueryResult ...
type QueryResult struct {
	Key    string `json:"Key"`
	Record *Patient
}

// CreatePatient ...
func (s *SimpleContract) CreatePatient(ctx contractapi.TransactionContextInterface, id string, name string, preExistingConditions string, diagnosisID string, statusID string, keyID string) error {
	patient := Patient{
		Name:                  name,
		PreExistingConditions: preExistingConditions,
		DiagnosisID:           diagnosisID,
		StatusID:              statusID,
		KeyID:                 keyID,
	}

	patientAsBytes, _ := json.Marshal(patient)

	return ctx.GetStub().PutState(id, patientAsBytes)
}

// FindPatient ...
func (s *SimpleContract) FindPatient(ctx contractapi.TransactionContextInterface, id string) (*Patient, error) {
	patientAsBytes, err := ctx.GetStub().GetState(id)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if patientAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", id)
	}

	patient := new(Patient)
	_ = json.Unmarshal(patientAsBytes, patient)

	return patient, nil
}

// AllPatients ...
func (s *SimpleContract) AllPatients(ctx contractapi.TransactionContextInterface, firstID string, lastID string) ([]QueryResult, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(firstID, lastID)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		patient := new(Patient)
		_ = json.Unmarshal(queryResponse.Value, patient)

		queryResult := QueryResult{Key: queryResponse.Key, Record: patient}
		results = append(results, queryResult)
	}

	return results, nil
}

// UpdatePatient ...
func (s *SimpleContract) UpdatePatient(ctx contractapi.TransactionContextInterface, id string, name string, preExistingConditions string, diagnosisID string, statusID string, keyID string) error {
	patient, err := s.FindPatient(ctx, id)

	if err != nil {
		return err
	}

	patient.Name = name
	patient.PreExistingConditions = preExistingConditions
	patient.DiagnosisID = diagnosisID
	patient.StatusID = statusID
	patient.KeyID = keyID

	patientAsBytes, _ := json.Marshal(patient)

	return ctx.GetStub().PutState(id, patientAsBytes)
}

// CreateProposal ...
func (s *SimpleContract) CreateProposal(ctx contractapi.TransactionContextInterface, id string, requesterID string, requestedID string, patientsIDs string, keyID string, modulo string) error {
	var ms []string

	proposal := Proposal{
		RequesterID: requesterID,
		RequestedID: requestedID,
		PatientsIDs: patientsIDs,
		KeyID:       keyID,
	}

	// Split patients' ids
	pids := strings.Split(proposal.PatientsIDs, ",")

	// Get all patients' values
	for _, pid := range pids {
		patient, err := s.FindPatient(ctx, pid)

		if err != nil {
			return err
		}

		ms = append(ms, patient.PreExistingConditions)
	}

	// Calculate average
	m := phe.MeanFromString(modulo, ms)

	// Save proposal
	proposal.Value = m
	proposalAsBytes, _ := json.Marshal(proposal)

	return ctx.GetStub().PutState(id, proposalAsBytes)
}

// FindProposal ...
func (s *SimpleContract) FindProposal(ctx contractapi.TransactionContextInterface, id string) (*Proposal, error) {
	proposalAsBytes, err := ctx.GetStub().GetState(id)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if proposalAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", id)
	}

	proposal := new(Proposal)
	_ = json.Unmarshal(proposalAsBytes, proposal)

	return proposal, nil
}

// CreateResult ...
func (s *SimpleContract) CreateResult(ctx contractapi.TransactionContextInterface, proposalID string, firstToken string, secondToken string, keyID string, modulo string) error {
	proposal, err := s.FindProposal(ctx, proposalID)

	if err != nil {
		return err
	}

	newValue := phe.KeyUpdateFromString(modulo, firstToken, secondToken, proposal.Value)

	result := Result{
		ProposalID: proposalID,
		KeyID:      keyID,
		Value:      newValue,
	}

	// Get the number out of proposal ID
	re := regexp.MustCompile(`[0-9]+`)
	idNumber := string(re.Find([]byte(proposalID)))
	id := "RESULT" + idNumber

	resultAsBytes, _ := json.Marshal(result)

	return ctx.GetStub().PutState(id, resultAsBytes)
}

// FindResult ...
func (s *SimpleContract) FindResult(ctx contractapi.TransactionContextInterface, id string) (*Result, error) {
	resultAsBytes, err := ctx.GetStub().GetState(id)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if resultAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", id)
	}

	result := new(Result)
	_ = json.Unmarshal(resultAsBytes, result)

	return result, nil
}
