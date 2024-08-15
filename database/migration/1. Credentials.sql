drop table if exists credentials;

CREATE TABLE public.credentials (
	id uuid NOT NULL,
	user_id uuid NOT NULL,
	service varchar NOT NULL,
	"type" varchar NOT NULL,
	"data" jsonb NOT NULL,
	updated_at timestamp NULL DEFAULT now(),
	created_at timestamp NULL DEFAULT now(),
    CONSTRAINT credentials_pk PRIMARY KEY (id),
    CONSTRAINT credentials_unique UNIQUE (user_id,service)
);
