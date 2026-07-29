-- One-time, idempotent migration from the legacy Xuancat Granter MySQL
-- database into NewAPI's integrated activation-code tables.
--
-- Run only after the new NewAPI binary has created activation_codes and
-- activation_logs. Both schemas must be reachable by the current MySQL user.

START TRANSACTION;

SET @default_owner_id := (
  SELECT user_id
  FROM `xuancat-granter`.newapi_keys
  WHERE is_default = TRUE AND is_active = TRUE
  ORDER BY id
  LIMIT 1
);

INSERT INTO `new-api`.activation_codes (
  user_id,
  code,
  days,
  channel,
  expired_time,
  status,
  created_time,
  used_time
)
SELECT
  COALESCE(k.user_id, @default_owner_id),
  c.code,
  c.days,
  c.channel,
  UNIX_TIMESTAMP(CONCAT(c.expire_date, ' 23:59:59')),
  1,
  UNIX_TIMESTAMP(c.created_at),
  0
FROM `xuancat-granter`.activation_codes AS c
LEFT JOIN `xuancat-granter`.newapi_keys AS k
  ON k.id = c.api_key_id
ON DUPLICATE KEY UPDATE code = VALUES(code);

INSERT INTO `new-api`.activation_codes (
  user_id,
  code,
  days,
  channel,
  expired_time,
  status,
  created_time,
  used_time
)
SELECT
  COALESCE(
    (
      SELECT t.user_id
      FROM `new-api`.tokens AS t
      WHERE t.`key` COLLATE utf8mb4_unicode_ci =
        (
          CASE
            WHEN latest.api_key LIKE 'sk-%' THEN SUBSTRING(latest.api_key, 4)
            ELSE latest.api_key
          END
        ) COLLATE utf8mb4_unicode_ci
      LIMIT 1
    ),
    (
      SELECT t.user_id
      FROM `new-api`.tokens AS t
      WHERE t.name COLLATE utf8mb4_unicode_ci =
        latest.identifier COLLATE utf8mb4_unicode_ci
      ORDER BY t.id DESC
      LIMIT 1
    ),
    @default_owner_id
  ),
  latest.activation_code,
  latest.days,
  'legacy',
  UNIX_TIMESTAMP(latest.used_at),
  2,
  UNIX_TIMESTAMP(latest.used_at),
  UNIX_TIMESTAMP(latest.used_at)
FROM `xuancat-granter`.activation_logs AS latest
INNER JOIN (
  SELECT activation_code, MAX(id) AS latest_id
  FROM `xuancat-granter`.activation_logs
  GROUP BY activation_code
) AS grouped
  ON grouped.latest_id = latest.id
ON DUPLICATE KEY UPDATE code = VALUES(code);

INSERT INTO `new-api`.activation_logs (
  legacy_source_id,
  activation_code_id,
  user_id,
  activation_code,
  action,
  days,
  identifier,
  token_id,
  token_key,
  client_ip,
  used_time
)
SELECT
  l.id,
  c.id,
  c.user_id,
  l.activation_code,
  l.action,
  l.days,
  l.identifier,
  COALESCE(
    (
      SELECT t.id
      FROM `new-api`.tokens AS t
      WHERE t.`key` COLLATE utf8mb4_unicode_ci =
        (
          CASE
            WHEN l.api_key LIKE 'sk-%' THEN SUBSTRING(l.api_key, 4)
            ELSE l.api_key
          END
        ) COLLATE utf8mb4_unicode_ci
      LIMIT 1
    ),
    (
      SELECT t.id
      FROM `new-api`.tokens AS t
      WHERE t.name COLLATE utf8mb4_unicode_ci =
        l.identifier COLLATE utf8mb4_unicode_ci
      ORDER BY t.id DESC
      LIMIT 1
    ),
    0
  ),
  CASE
    WHEN l.api_key = '' THEN COALESCE(
      (
        SELECT CONCAT('sk-', t.`key`)
        FROM `new-api`.tokens AS t
        WHERE t.name COLLATE utf8mb4_unicode_ci =
          l.identifier COLLATE utf8mb4_unicode_ci
        ORDER BY t.id DESC
        LIMIT 1
      ),
      ''
    )
    ELSE l.api_key
  END,
  COALESCE(l.client_ip, ''),
  UNIX_TIMESTAMP(l.used_at)
FROM `xuancat-granter`.activation_logs AS l
INNER JOIN `new-api`.activation_codes AS c
  ON c.code COLLATE utf8mb4_unicode_ci =
    l.activation_code COLLATE utf8mb4_unicode_ci
ON DUPLICATE KEY UPDATE legacy_source_id = VALUES(legacy_source_id);

COMMIT;
