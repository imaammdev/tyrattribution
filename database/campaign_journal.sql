CREATE TABLE campaign_journal (
    campaign_journal_id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    campaign_id UUID NOT NULL,
    date DATE NOT NULL,
    number_of_click BIGINT,
    number_of_conversion BIGINT,
    total_conversion_value DECIMAL(10,2),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);