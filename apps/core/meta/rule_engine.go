package meta

import (
	"fmt"
	"github.com/goerp/goerp/apps/core/database"
	"github.com/goerp/goerp/apps/core/logger"
	"github.com/yuin/gopher-lua"
	"go.uber.org/zap"
)

type RuleEngine struct {
	L *lua.LState
}

func NewRuleEngine() *RuleEngine {
	return &RuleEngine{L: lua.NewState()}
}

func (e *RuleEngine) Close() {
	e.L.Close()
}

// EvaluateCondition checks if a rule should be applied based on document data
func (e *RuleEngine) EvaluateCondition(conditionScript string, doc map[string]interface{}) (bool, error) {
	L := e.L
	
	// Create Lua table from doc
	lt := L.NewTable()
	for k, v := range doc {
		L.SetTable(lt, lua.LString(k), lua.LString(fmt.Sprintf("%v", v)))
	}
	L.SetGlobal("doc", lt)

	// Execute condition script
	if err := L.DoString("return " + conditionScript); err != nil {
		return false, err
	}

	res := L.Get(-1)
	L.Pop(1)
	return lua.LVAsBool(res), nil
}

// ApplyRule executes the action script on the document
func (e *RuleEngine) ApplyAction(actionScript string, doc map[string]interface{}) (map[string]interface{}, error) {
	L := e.L
	
	lt := L.NewTable()
	for k, v := range doc {
		L.SetTable(lt, lua.LString(k), lua.LString(fmt.Sprintf("%v", v)))
	}
	L.SetGlobal("doc", lt)

	if err := L.DoString(actionScript); err != nil {
		return nil, err
	}

	// Read modified doc back from Lua
	updatedDoc := make(map[string]interface{})
	L.GetGlobal("doc").(*lua.LTable).ForEach(func(k, v lua.LValue) {
		updatedDoc[k.String()] = v.String()
	})

	return updatedDoc, nil
}

// ProcessRules fetches and executes all applicable rules for a DocType
func ProcessRules(doctype string, doc map[string]interface{}, tenantID string) (map[string]interface{}, error) {
	var rules []map[string]interface{}
	// Fetch active rules from database (e.g., tabPricingRule or tabBusinessRule)
	database.DB.Table("tabBusinessRule").
		Where("target_doctype = ? AND is_active = ? AND tenant_id = ?", doctype, true, tenantID).
		Order("priority DESC").
		Find(&rules)

	if len(rules) == 0 {
		return doc, nil
	}

	engine := NewRuleEngine()
	defer engine.Close()

	currentDoc := doc
	for _, rule := range rules {
		condition := rule["condition"].(string)
		action := rule["action_script"].(string)

		match, err := engine.EvaluateCondition(condition, currentDoc)
		if err != nil {
			logger.Log.Error("Rule condition error", zap.Error(err), zap.String("rule", rule["name"].(string)))
			continue
		}

		if match {
			logger.Log.Info("Applying rule", zap.String("rule", rule["name"].(string)))
			updated, err := engine.ApplyAction(action, currentDoc)
			if err != nil {
				logger.Log.Error("Rule action error", zap.Error(err))
				continue
			}
			currentDoc = updated
		}
	}

	return currentDoc, nil
}
