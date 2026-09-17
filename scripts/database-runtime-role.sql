-- Operator-run, NOT automatic migration. Apply as the schema owner, test in staging.
-- Provision password separately with psql \password or the provider secret workflow.
BEGIN;
DO $setup$
BEGIN
 IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname='prepyo_api') THEN
  CREATE ROLE prepyo_api LOGIN NOSUPERUSER NOCREATEDB NOCREATEROLE NOREPLICATION NOBYPASSRLS;
 END IF;
 IF EXISTS (SELECT 1 FROM pg_roles WHERE rolname='prepyo_api' AND
  (rolsuper OR rolcreatedb OR rolcreaterole OR rolreplication OR rolbypassrls)) THEN
  RAISE EXCEPTION 'prepyo_api already has elevated privileges; review before rollout';
 END IF;
END $setup$;
GRANT USAGE ON SCHEMA public TO prepyo_api;
DO $grants$
DECLARE tab record;
BEGIN
 FOR tab IN SELECT c.relname, c.relrowsecurity FROM pg_class c
 JOIN pg_namespace n ON n.oid=c.relnamespace
 WHERE n.nspname='public' AND c.relkind IN ('r','p') AND c.relname<>'schema_migrations'
 LOOP
  EXECUTE format('GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.%I TO prepyo_api', tab.relname);
  IF tab.relrowsecurity AND NOT EXISTS (
   SELECT 1 FROM pg_policies WHERE schemaname='public' AND tablename=tab.relname AND policyname='prepyo_server_access'
  ) THEN
   EXECUTE format('CREATE POLICY prepyo_server_access ON public.%I FOR ALL TO prepyo_api USING (true) WITH CHECK (true)', tab.relname);
  END IF;
 END LOOP;
END $grants$;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO prepyo_api;
COMMIT;
