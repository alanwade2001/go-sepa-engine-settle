package service

import (
	"log/slog"
	"time"

	"github.com/alanwade2001/go-sepa-engine-data/model"
	"github.com/alanwade2001/go-sepa-engine-data/repository"
	"github.com/alanwade2001/go-sepa-engine-data/repository/entity"
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
			perSg.CtrlSum = amount
			perSg.NbOfTxs = uint(rows)
			perSg.CreDtTm = &now
			perSg.MsgID = "pacs008-1"

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
