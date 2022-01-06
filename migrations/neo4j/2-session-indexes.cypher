CREATE CONSTRAINT session_id_unique IF NOT EXISTS
FOR (s:Session)
REQUIRE s.uuid IS UNIQUE;

CREATE INDEX session_expires_index IF NOT EXISTS
FOR (s:Session)
ON (s.expires);
