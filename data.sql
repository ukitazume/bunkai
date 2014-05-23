
-- ----------------------------
--  Sequence structure for sentences_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "public"."sentences_id_seq";
CREATE SEQUENCE "public"."sentences_id_seq" INCREMENT 1 START 5 MAXVALUE 9223372036854775807 MINVALUE 1 CACHE 1;

-- ----------------------------
--  Table structure for sentences
-- ----------------------------
DROP TABLE IF EXISTS "public"."sentences";
CREATE TABLE "public"."sentences" (
	"id" int4 NOT NULL DEFAULT nextval('sentences_id_seq'::regclass),
	"text" varchar(255) NOT NULL COLLATE "default",
	"url" varchar(2083) NOT NULL COLLATE "default",
	"created_at" timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
)
WITH (OIDS=FALSE);


-- -- ----------------------------
-- --  Alter sequences owned by
-- -- ----------------------------
-- ALTER SEQUENCE "public"."sentences_id_seq" RESTART 4 OWNED BY "sentences"."id";
-- 
-- -- ----------------------------
-- --  Primary key structure for table sentences
-- -- ----------------------------
-- ALTER TABLE "public"."sentences" ADD PRIMARY KEY ("id") NOT DEFERRABLE INITIALLY IMMEDIATE;
-- 
-- 
