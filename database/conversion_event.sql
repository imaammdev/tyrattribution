CREATE TABLE conversion_event (
    conversion_id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID NOT NULL,
    campaign_id UUID NOT NULL,
    click_id UUID,
    conversion_date TIMESTAMP NOT NULL,
    value DECIMAL(10,2),
    type VARCHAR(255) NOT NULL,
    source VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_conversion_event_campaign ON conversion_event (campaign_id);