CREATE DATABASE currency_exchange;
GO

USE currency_exchange;
GO

CREATE TABLE conversion_history (
    id INT IDENTITY(1,1) PRIMARY KEY,
    from_currency VARCHAR(3) NOT NULL,
    to_currency VARCHAR(3) NOT NULL,
    amount DECIMAL(18,4) NOT NULL,
    result DECIMAL(18,4) NOT NULL,
    rate DECIMAL(18,6) NOT NULL,
    created_at DATETIME2 DEFAULT GETDATE()
);
GO

CREATE INDEX idx_history_date ON conversion_history(created_at);
CREATE INDEX idx_history_currencies ON conversion_history(from_currency, to_currency);
GO

CREATE OR ALTER PROCEDURE sp_InsertConversion
    @from_currency VARCHAR(3),
    @to_currency VARCHAR(3),
    @amount DECIMAL(18,4),
    @result DECIMAL(18,4),
    @rate DECIMAL(18,6)
AS
BEGIN
    SET NOCOUNT ON;
    INSERT INTO conversion_history (from_currency, to_currency, amount, result, rate)
    VALUES (@from_currency, @to_currency, @amount, @result, @rate);
    
    SELECT SCOPE_IDENTITY() AS id;
END;
GO

CREATE OR ALTER PROCEDURE sp_GetConversionHistory
    @from_currency VARCHAR(3) = NULL,
    @to_currency VARCHAR(3) = NULL,
    @limit INT = 50
AS
BEGIN
    SET NOCOUNT ON;
    SELECT TOP(@limit) id, from_currency, to_currency, amount, result, rate, created_at
    FROM conversion_history
    WHERE (@from_currency IS NULL OR from_currency = @from_currency)
      AND (@to_currency IS NULL OR to_currency = @to_currency)
    ORDER BY created_at DESC;
END;
GO
CREATE OR ALTER PROCEDURE sp_GetConversionsByDateRange
    @start_date DATETIME2,
    @end_date DATETIME2
AS
BEGIN
    SET NOCOUNT ON;
    SELECT id, from_currency, to_currency, amount, result, rate, created_at
    FROM conversion_history
    WHERE created_at BETWEEN @start_date AND @end_date
    ORDER BY created_at DESC;
END;
GO
