CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  email varchar(255) NOT NULL,
  name text,
  status varchar(20) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email ON users (email);

CREATE TABLE IF NOT EXISTS credentials (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  user_id uuid NOT NULL,
  password_hash varchar(255) NOT NULL,
  CONSTRAINT chk_credentials_password_hash_length CHECK (char_length(password_hash) >= 8),
  CONSTRAINT fk_credentials_user FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_credentials_user_id ON credentials (user_id);

CREATE TABLE IF NOT EXISTS refresh_tokens (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  user_id uuid NOT NULL,
  token varchar(64) NOT NULL,
  expires_at timestamptz NOT NULL,
  revoked boolean NOT NULL DEFAULT false,
  CONSTRAINT fk_refresh_tokens_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_refresh_tokens_token ON refresh_tokens (token);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens (user_id);

CREATE TABLE IF NOT EXISTS friendships (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  user_id uuid NOT NULL,
  friend_id uuid NOT NULL,
  CONSTRAINT fk_friendships_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_friendships_friend FOREIGN KEY (friend_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_user_friend ON friendships (user_id, friend_id);
CREATE INDEX IF NOT EXISTS idx_friendships_user_id ON friendships (user_id);
CREATE INDEX IF NOT EXISTS idx_friendships_friend_id ON friendships (friend_id);

CREATE TABLE IF NOT EXISTS friend_requests (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  sender_id uuid NOT NULL,
  receiver_id uuid NOT NULL,
  status varchar(20) NOT NULL,
  idempotency_key varchar(64),
  CONSTRAINT fk_friend_requests_sender FOREIGN KEY (sender_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_friend_requests_receiver FOREIGN KEY (receiver_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_friend_requests_idempotency_key ON friend_requests (idempotency_key);
CREATE INDEX IF NOT EXISTS idx_friend_requests_sender_id ON friend_requests (sender_id);
CREATE INDEX IF NOT EXISTS idx_friend_requests_receiver_id ON friend_requests (receiver_id);

CREATE TABLE IF NOT EXISTS groups (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  name varchar(100) NOT NULL,
  owner_id uuid NOT NULL,
  simplify_debts boolean NOT NULL DEFAULT false,
  CONSTRAINT fk_groups_owner FOREIGN KEY (owner_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_groups_deleted_at ON groups (deleted_at);

CREATE TABLE IF NOT EXISTS group_memberships (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  user_id uuid NOT NULL,
  group_id uuid NOT NULL,
  role varchar(20) NOT NULL,
  CONSTRAINT fk_group_memberships_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_group_memberships_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_group_user ON group_memberships (user_id, group_id);
CREATE INDEX IF NOT EXISTS idx_group_memberships_user_id ON group_memberships (user_id);
CREATE INDEX IF NOT EXISTS idx_group_memberships_group_id ON group_memberships (group_id);

CREATE TABLE IF NOT EXISTS splits (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  type varchar(20) NOT NULL,
  division_type varchar(20) NOT NULL,
  total_amount bigint NOT NULL,
  currency varchar(10) NOT NULL,
  description varchar(500),
  simplify_debts boolean DEFAULT NULL,
  idempotency_key varchar(64),
  group_id uuid,
  created_by_id uuid NOT NULL,
  CONSTRAINT fk_splits_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE SET NULL,
  CONSTRAINT fk_splits_created_by FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_splits_idempotency_key ON splits (idempotency_key);
CREATE INDEX IF NOT EXISTS idx_splits_group_id ON splits (group_id);
CREATE INDEX IF NOT EXISTS idx_splits_created_by_id ON splits (created_by_id);

CREATE TABLE IF NOT EXISTS split_participants (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  split_id uuid NOT NULL,
  user_id uuid NOT NULL,
  share_amount bigint NOT NULL,
  settled_amount bigint NOT NULL DEFAULT 0,
  currency varchar(10) NOT NULL,
  is_settled boolean NOT NULL DEFAULT false,
  CONSTRAINT fk_split_participants_split FOREIGN KEY (split_id) REFERENCES splits(id) ON DELETE CASCADE,
  CONSTRAINT fk_split_participants_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_split_user ON split_participants (split_id, user_id);
CREATE INDEX IF NOT EXISTS idx_split_participants_split_id ON split_participants (split_id);
CREATE INDEX IF NOT EXISTS idx_split_participants_user_id ON split_participants (user_id);

CREATE TABLE IF NOT EXISTS settlements (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  split_id uuid NOT NULL,
  payer_id uuid NOT NULL,
  payee_id uuid NOT NULL,
  amount bigint NOT NULL,
  currency varchar(10) NOT NULL,
  date timestamptz NOT NULL,
  confirmed boolean NOT NULL DEFAULT false,
  idempotency_key varchar(64),
  CONSTRAINT fk_settlements_split FOREIGN KEY (split_id) REFERENCES splits(id) ON DELETE CASCADE,
  CONSTRAINT fk_settlements_payer FOREIGN KEY (payer_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_settlements_payee FOREIGN KEY (payee_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_settlements_idempotency_key ON settlements (idempotency_key);
CREATE INDEX IF NOT EXISTS idx_settlements_split_id ON settlements (split_id);
CREATE INDEX IF NOT EXISTS idx_settlements_payer_id ON settlements (payer_id);
CREATE INDEX IF NOT EXISTS idx_settlements_payee_id ON settlements (payee_id);

CREATE TABLE IF NOT EXISTS user_balances (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  user_id uuid NOT NULL,
  other_user_id uuid NOT NULL,
  net_amount bigint NOT NULL DEFAULT 0,
  currency varchar(10) NOT NULL,
  CONSTRAINT fk_user_balances_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_user_balances_other_user FOREIGN KEY (other_user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_user_balances_user_id ON user_balances (user_id);
CREATE INDEX IF NOT EXISTS idx_user_balances_other_user_id ON user_balances (other_user_id);

CREATE TABLE IF NOT EXISTS group_balances (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  deleted_at timestamptz,
  user_id uuid NOT NULL,
  group_id uuid NOT NULL,
  net_amount bigint NOT NULL DEFAULT 0,
  currency varchar(10) NOT NULL,
  CONSTRAINT fk_group_balances_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT fk_group_balances_group FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_group_balances_user_id ON group_balances (user_id);
CREATE INDEX IF NOT EXISTS idx_group_balances_group_id ON group_balances (group_id);
