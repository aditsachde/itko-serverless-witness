CREATE TABLE chkpts (
    region VARCHAR NOT NULL,
    log_id VARCHAR NOT NULL,
    chkpt BYTEA,
    range BYTEA,
    PRIMARY KEY (region, log_id)
);
