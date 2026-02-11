CREATE TABLE IF NOT EXISTS subscriptions (
    id UUID PRIMARY KEY,
    service_name TEXT NOT NULL,
    month_cost INTEGER NOT NULL,
    user_id UUID NOT NULL,
    subs_start_date DATE NOT NULL DEFAULT CURRENT_DATE,
    subs_end_date DATE
);