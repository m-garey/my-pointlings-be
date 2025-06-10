-- WARNING: This schema is for context only and is not meant to be run.
-- Table order and constraints may not be valid for execution.

CREATE TABLE public.items (
  item_id bigint NOT NULL DEFAULT nextval('items_item_id_seq'::regclass),
  item_category USER-DEFINED NOT NULL,
  slot USER-DEFINED,
  asset_id text NOT NULL,
  name text NOT NULL,
  rarity USER-DEFINED NOT NULL,
  price_points integer,
  unlock_level integer,
  CONSTRAINT items_pkey PRIMARY KEY (item_id)
);
CREATE TABLE public.point_spend (
  spend_id bigint NOT NULL DEFAULT nextval('point_spend_spend_id_seq'::regclass),
  user_id bigint NOT NULL,
  item_id bigint NOT NULL,
  points_spent integer NOT NULL CHECK (points_spent > 0),
  spend_ts timestamp with time zone NOT NULL DEFAULT now(),
  CONSTRAINT point_spend_pkey PRIMARY KEY (spend_id),
  CONSTRAINT point_spend_item_id_fkey FOREIGN KEY (item_id) REFERENCES public.items(item_id),
  CONSTRAINT point_spend_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id)
);
CREATE TABLE public.pointling_colors (
  pointling_id bigint NOT NULL,
  color_hex character NOT NULL,
  acquired_at timestamp with time zone NOT NULL DEFAULT now(),
  CONSTRAINT pointling_colors_pkey PRIMARY KEY (pointling_id, color_hex),
  CONSTRAINT pointling_colors_pointling_id_fkey FOREIGN KEY (pointling_id) REFERENCES public.pointlings(pointling_id)
);
CREATE TABLE public.pointling_items (
  pointling_id bigint NOT NULL,
  item_id bigint NOT NULL,
  acquired_at timestamp with time zone NOT NULL DEFAULT now(),
  equipped boolean NOT NULL DEFAULT false,
  CONSTRAINT pointling_items_pkey PRIMARY KEY (pointling_id, item_id),
  CONSTRAINT pointling_items_pointling_id_fkey FOREIGN KEY (pointling_id) REFERENCES public.pointlings(pointling_id),
  CONSTRAINT pointling_items_item_id_fkey FOREIGN KEY (item_id) REFERENCES public.items(item_id)
);
CREATE TABLE public.pointlings (
  pointling_id bigint NOT NULL DEFAULT nextval('pointlings_pointling_id_seq'::regclass),
  user_id bigint NOT NULL,
  nickname text,
  level integer NOT NULL DEFAULT 1,
  current_xp integer NOT NULL DEFAULT 0,
  required_xp integer NOT NULL DEFAULT 3,
  personality_id integer,
  look_json jsonb NOT NULL DEFAULT '{}'::jsonb,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  CONSTRAINT pointlings_pkey PRIMARY KEY (pointling_id),
  CONSTRAINT pointlings_user_id_fkey FOREIGN KEY (user_id) REFERENCES public.users(user_id)
);
CREATE TABLE public.users (
  user_id bigint NOT NULL,
  display_name text NOT NULL,
  point_balance bigint NOT NULL DEFAULT 0,
  created_at timestamp with time zone NOT NULL DEFAULT now(),
  CONSTRAINT users_pkey PRIMARY KEY (user_id)
);
CREATE TABLE public.xp_events (
  event_id bigint NOT NULL DEFAULT nextval('xp_events_event_id_seq'::regclass),
  pointling_id bigint NOT NULL,
  source USER-DEFINED NOT NULL,
  xp_amount integer NOT NULL,
  event_ts timestamp with time zone NOT NULL DEFAULT now(),
  CONSTRAINT xp_events_pkey PRIMARY KEY (event_id),
  CONSTRAINT xp_events_pointling_id_fkey FOREIGN KEY (pointling_id) REFERENCES public.pointlings(pointling_id)
);
