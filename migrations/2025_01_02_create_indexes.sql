-- Create indexes on necessary fields
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_transactions_from_user_id ON transactions(from_user_id);
CREATE INDEX idx_transactions_to_user_id ON transactions(to_user_id);
CREATE INDEX idx_balances_user_id ON balances(user_id);
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity_type, entity_id);
