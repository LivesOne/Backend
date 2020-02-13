CREATE TABLE `livesone_assert`.`user_asset` (
  `uid` BIGINT NOT NULL,
  `balance` BIGINT NOT NULL DEFAULT 0,
  `nonce` BIGINT NOT NULL DEFAULT 1,
  `lastmodify` INT UNSIGNED NOT NULL DEFAULT 0,
  `status` SMALLINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`uid`),
  INDEX `BALANCE` (`balance`),
  INDEX `LASTMODIFY` (`lastmodify`),
  INDEX `STATUS` (`status`)
 ) 
ENGINE=InnoDB
DEFAULT CHARACTER SET utf8;

CREATE TABLE `livesone_assert`.`user_reward` (
  `uid` BIGINT UNSIGNED NOT NULL,
  `total` BIGINT NOT NULL DEFAULT 0,
  `lastday` INT NOT NULL DEFAULT 0,
  `lastmodify` INT UNSIGNED NOT NULL DEFAULT 0,
  PRIMARY KEY (`uid`),
  INDEX `TOTAL` (`total`),
  INDEX `LASTMODIFY` (`lastmodify`)
)
ENGINE=InnoDB
DEFAULT CHARACTER SET utf8;

-- CREATE TABLE `livesone_assert`.`miner_history` (
--   `txid` BIGINT UNSIGNED NOT NULL,
--   `uid` BIGINT NOT NULL,
--   `value` INT NOT NULL DEFAULT 0,
--   `date` DATE NOT NULL,
--   `channel` SMALLINT NOT NULL DEFAULT 0,
--   `subchannel` SMALLINT NOT NULL DEFAULT 0,
--   `time` INT UNSIGNED NOT NULL DEFAULT 0,
--   PRIMARY KEY (`txid`),
--   INDEX `UID` (`uid`),
--   INDEX `TIME` (`time`),
--   INDEX `CHANNEL` (`channel`),
--   INDEX `SUBCHANNEL` (`subchannel`)
-- )
-- ENGINE=InnoDB
-- DEFAULT CHARACTER SET utf8;

CREATE TABLE `livesone_assert`.`recent_tx_ids` (
  `txid` BIGINT UNSIGNED NOT NULL,
  PRIMARY KEY (`txid`)
)
ENGINE=InnoDB
DEFAULT CHARACTER SET utf8;

-- CREATE TABLE `livesone_assert`.`transfer_history` (
--   `txid` BIGINT UNSIGNED NOT NULL,
--   `from` BIGINT NOT NULL,
--   `from_nonce` BIGINT NOT NULL,
--   `from_balance` BIGINT NOT NULL,
--   `to` BIGINT NOT NULL,
--   `to_balance` BIGINT NOT NULL,
--   `value` INT NOT NULL,  
--   `time` INT UNSIGNED NOT NULL DEFAULT 0,
--   `code` SMALLINT NOT NULL DEFAULT 0,
--   PRIMARY KEY (`txid`),
--   INDEX `FROM` (`from`),
--   INDEX `TO` (`to`),
--   INDEX `TIME` (`time`),
--   INDEX `CODE` (`code`)
-- )
-- ENGINE=InnoDB
-- DEFAULT CHARACTER SET utf8;

CREATE TABLE `livesone_assert`.`livesone_asset` (
  `id` INT NOT NULL,
  `balance` BIGINT NOT NULL DEFAULT 0,
  PRIMARY KEY (`id`)
)
ENGINE=InnoDB
DEFAULT CHARACTER SET utf8;

-- mongo:
-- tx_history: {
--   txid: primary key
--   type:
--   from:
--   to:
--   value:
--   ts:
--   code:
--   miner: {
--     chn:
--   }
--   trans: {
--     nonce:    
--   }
-- }
-- _id: txid
-- Indexs:
-- 1.from
-- 2.to

