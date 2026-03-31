-- Remove alert_type column from budget_alerts (rollback)
ALTER TABLE budget_alerts DROP COLUMN IF EXISTS alert_type;

-- Drop budget_alerts table
DROP TABLE IF EXISTS budget_alerts;

-- Drop budgets table
DROP TABLE IF EXISTS budgets;
