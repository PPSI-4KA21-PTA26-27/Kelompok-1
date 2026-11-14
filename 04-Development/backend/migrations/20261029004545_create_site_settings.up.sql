CREATE TABLE site_settings (
    id                  INT NOT NULL,
    maintenance_mode    BOOLEAN NOT NULL DEFAULT FALSE,
    maintenance_message TEXT NOT NULL,

    CONSTRAINT pk_site_settings PRIMARY KEY (id)
);
