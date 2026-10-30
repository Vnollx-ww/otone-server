CREATE TABLE users (
    id          CHAR(36)     NOT NULL,
    device_id   VARCHAR(128) NOT NULL,
    platform    VARCHAR(32)  NOT NULL DEFAULT '',
    status      TINYINT      NOT NULL DEFAULT 1 COMMENT '1=active, 2=disabled',
    created_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_users_device (device_id, platform)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE players (
    user_id     CHAR(36)    NOT NULL,
    nickname    VARCHAR(32) NOT NULL DEFAULT '',
    level       INT         NOT NULL DEFAULT 1,
    exp         BIGINT      NOT NULL DEFAULT 0,
    created_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    updated_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (user_id),
    CONSTRAINT fk_players_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE refresh_tokens (
    id          CHAR(36)    NOT NULL,
    user_id     CHAR(36)    NOT NULL,
    token_hash  CHAR(64)    NOT NULL,
    expires_at  DATETIME(3) NOT NULL,
    revoked_at  DATETIME(3) NULL,
    created_at  DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    PRIMARY KEY (id),
    UNIQUE KEY uk_refresh_token_hash (token_hash),
    KEY idx_refresh_user_expires (user_id, expires_at),
    CONSTRAINT fk_refresh_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

CREATE TABLE story_progress (
    user_id          CHAR(36)    NOT NULL,
    chapter_id       VARCHAR(64) NOT NULL,
    node_id          VARCHAR(64) NOT NULL,
    content_version  VARCHAR(32) NOT NULL DEFAULT '',
    flags            JSON        NOT NULL,
    updated_at       DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (user_id, chapter_id),
    KEY idx_story_user_updated (user_id, updated_at),
    CONSTRAINT fk_story_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
