CREATE TABLE click_event (
    click_id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    campaign_id UUID NOT NULL,
    user_id UUID NOT NULL,
    click_date TIMESTAMP NOT NULL,
    source VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

CREATE INDEX idx_click_event_composite ON click_event (campaign_id, user_id, source, click_date);