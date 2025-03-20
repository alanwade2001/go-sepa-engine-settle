package service

import (
	"encoding/xml"
	"log/slog"
	"strconv"
	"time"

	"github.com/alanwade2001/go-sepa-engine-data/model"
	"github.com/alanwade2001/go-sepa-engine-data/repository"
	"github.com/alanwade2001/go-sepa-engine-data/repository/entity"
	"github.com/alanwade2001/go-sepa-iso/pacs_008_001_02"
)

type Execution struct {
	reposMgr *repository.Manager
}

func NewExecution(reposMgr *repository.Manager) *Execution {
	document := &Execution{
		reposMgr: reposMgr,
	}

	return document
}

func (d *Execution) Execute(mdl *model.Execution) (*model.Execution, error) {

	if newExec, err := mdl.ToEntity(); err != nil {
		return nil, err
	} else if persisted, err := d.reposMgr.Execution.Perist(newExec); err != nil {
		return nil, err
	} else {

		newSg := &entity.SettlementGroup{
			Execution: persisted,
		}

		if perSg, err := d.reposMgr.SettlementGroup.Perist(newSg); err != nil {
			return nil, err
		} else if rows, err := d.reposMgr.Settlement.UpdateSettlementGroup(perSg); err != nil {
			return nil, err
		} else if amount, err := d.reposMgr.Settlement.SumSettlementAmountBySettlementGroupID(perSg.Model.ID); err != nil {
			return nil, err
		} else {
			now := time.Now()
			gh := &pacs_008_001_02.GroupHeader33{
				MsgId:   "SCTORD156820211213000000012649",
				CreDtTm: now.Format("2006-12-13T13:24:39"),
				NbOfTxs: strconv.FormatInt(rows, 10),
				CtrlSum: amount,
				TtlIntrBkSttlmAmt: &pacs_008_001_02.ActiveCurrencyAndAmount{
					CcyAttr: "EUR",
					Value:   amount,
				},
				SttlmInf: &pacs_008_001_02.SettlementInformation13{
					SttlmMtd: "CLRG",
					ClrSys: &pacs_008_001_02.ClearingSystemIdentification3Choice{
						Prtry: "ST2",
					},
				},
				InstgAgt: &pacs_008_001_02.BranchAndFinancialInstitutionIdentification4{
					FinInstnId: &pacs_008_001_02.FinancialInstitutionIdentification7{
						BIC: "BTRLRO22",
					},
				},
				InstdAgt: &pacs_008_001_02.BranchAndFinancialInstitutionIdentification4{
					FinInstnId: &pacs_008_001_02.FinancialInstitutionIdentification7{
						BIC: "BTRLRO22",
					},
				},
			}

			if ghBytes, err := xml.Marshal(gh); err != nil {
				return nil, err
			} else {

				perSg.CtrlSum = amount
				perSg.NbOfTxs = uint(rows)
				perSg.CreDtTm = &now
				perSg.MsgID = "pacs008-1"
				perSg.GrpHdr = string(ghBytes)

				if updSg, err := d.reposMgr.SettlementGroup.Perist(perSg); err != nil {
					return nil, err
				} else {

					slog.Info("settlement group", "id", updSg.Model.ID, "nbOfTxs", updSg.NbOfTxs, "ctrlSum", updSg.CtrlSum)

					newMdl := &model.Execution{}

					if err := newMdl.FromEntity(persisted, []*entity.SettlementGroup{updSg}); err != nil {
						return nil, err
					} else {
						return newMdl, nil
					}
				}
			}
		}
	}

}
