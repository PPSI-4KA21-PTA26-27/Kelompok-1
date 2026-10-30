CREATE TABLE site_settings (
    id                  INT PRIMARY KEY,
    maintenance_mode    BOOLEAN NOT NULL DEFAULT FALSE,
    maintenance_message TEXT NOT NULL
);
