-- Add new tests
CREATE OR REPLACE FUNCTION web_backend__routing.f_callingsys_request_add(i_user_sessionid integer, i_rtype integer, i_callscount integer, s_requestname character varying, a_typeid_array integer[], j_recordset_as_json json, OUT status character varying, OUT message character varying)
 RETURNS record
 LANGUAGE plpgsql
AS $function$
DECLARE 
	i_count			integer;
	blsplit_task	boolean = False;
BEGIN
	status	:= 'FAILED';
	message	:= 'Invalid Test Parameters!';
	IF (COALESCE(j_recordset_as_json::TEXT, '') = '') THEN
		message	:= message || ': Request is empty';
		RETURN;
	ELSEIF (a_typeid_array IS NULL) OR (array_length(a_typeid_array, 1) = 0) THEN
		message	:= message || ': Incorrect Type';
		RETURN;
	ELSEIF COALESCE(s_requestname, '') = '' THEN
		message	:= message || ': Incorrect Request Name';
		RETURN;
	-- Check Name Dulicate	
	ELSEIF EXISTS (SELECT 1 FROM mtcarrierdbret."Purch_Oppt" 
				   	WHERE "RequestName" ~* FORMAT('^%s\(p_\d*\)', s_requestname) LIMIT 1) THEN
		message	:= message || ': Request Name is duplicated';
		RETURN;		
	ELSEIF COALESCE(i_callscount, 0) <= 0 THEN
		message	:= message || ': Incorrect Calls Count';
		RETURN;	
	END IF;	
	message	:= '';
	
	blsplit_task	:= (array_length(a_typeid_array, 1) > 1) OR (json_array_length(j_recordset_as_json) > 1);
	
	WITH testlist AS (SELECT 
							destinationid, 
							routeid,
							array_to_string(ARRAY(SELECT json_array_elements_text((bnumbers::json))), E'\r\n') bnumbers
						FROM json_to_recordset(j_recordset_as_json) AS x(destinationid integer, routeid integer, bnumbers text) 
				   ),
		test_types AS (SELECT tid FROM UNNEST(a_typeid_array) tid ),
		usr AS (SELECT web_backend.f_user_id_from_session_get(i_user_sessionid) AS usrid)
	INSERT INTO mtcarrierdbret."Purch_Oppt" ("Timezone", "Request_Date_Time", "Request_by_User", "Tested_by_User",
											"DestinationID", "Destination", 
											"RType", "Supplier", "SupplierID",
											"Test_Type", "RequestName", 
											 "Test_Calls", "Custom_BNumbers", "CallingSys_RouteID")
	SELECT 
			1, current_timestamp, usrid, usrid,
			destinationid, lpd."Destination", 
			i_rtype, lpc."Carrier", "Captura_CarrierID", 
			test_types.tid, 
			CASE 
				WHEN TRUE THEN 	FORMAT('%s(p_%s)', s_requestname, ROW_NUMBER () OVER (ORDER BY test_types.tid, lpd."Destination"))
				ELSE 			s_requestname
			END, 
			i_callscount, bnumbers, routeid
		FROM testlist tl
		JOIN mtcarrierdbret."Route_MasterDest_RType" lpd ON lpd."Destinationid" = tl.destinationid AND lpd."RType" =1 
		JOIN mtcarrierdbret."CallingSys_RouteList" rl ON rl."RouteID" = routeid
		JOIN mtcarrierdbret."LpCarrier" lpc ON lpc."Carrier_ID" = rl."Captura_CarrierID"
			, test_types
			, usr;
	GET DIAGNOSTICS i_count = ROW_COUNT;
		
	message	:= 'Tests added to que: ' || i_count;
	status	:= 'OK';		
END
$function$
;
---------------------------------------------------------------------------------------------------------
--Test results
CREATE OR REPLACE FUNCTION web_backend__routing.f_callingsys_request_calls_get(i_requestid integer)
 RETURNS TABLE(callid character varying, call_testid character varying, calltype character varying, destnumber character varying, callingnumber character varying, callstart timestamp without time zone, callend timestamp without time zone, duration double precision, pdd double precision, destname character varying, route character varying, result character varying, causecode integer, detected_callingnumber character varying, result_cli character varying, result_fas character varying, diagram bytea, recordid_audio bigint)
 LANGUAGE sql
AS $function$
	SELECT
		"CallID" callid,
		"CallListID" call_testid,
		"CallType" calltype,
		"BNumber" destnumber,
		"CallingNumber" callingnumber,
		"CallStart" callstart,
		"CallComplete" callend,
		"CallDuration" duration,
		"PDD" pdd,
		r."Destination" destname,
		"Route" route,
		"Status" result,
		"CauseCodeId" causecode,
		"CLIDetectedCallingNumber" detected_callingnumber,
		"CLIResult" result_cli,
		"FasResult" result_fas,
		cf.diagram,
		cf.record_id recordid
	FROM mtcarrierdbret."CallingSys_TestResults" r
		JOIN mtcarrierdbret."Purch_Oppt" po ON po."TestingSystemRequestID" = "CallListID"
		LEFT JOIN mtcarrierdbret."CallingSys_testfiles_web" cf ON cf.callid = r."CallID" AND cf.testsystem = r."TestSystem"
	WHERE "RequestID" = i_requestid
	ORDER BY "CallStart";
$function$
;
---------------------------------------------------------------------------------------------------------
--Tests list
CREATE OR REPLACE FUNCTION web_backend__routing.f_callingsys_request_list_get()
 RETURNS TABLE(reqid integer, stateid integer, statename character varying, reqname character varying, testtype character varying, destname character varying, supplier character varying, username character varying, req_dt timestamp without time zone, start_dt timestamp without time zone, end_dt timestamp without time zone, asr double precision, acd double precision, comp_calls bigint, incomp_calls bigint, minutes double precision, reqcomment text, reqresult character varying, callingsys_id character varying)
 LANGUAGE sql
AS $function$
	SELECT
			"RequestID" reqid,
			req.id, /*"RequestState" stateid,*/
			req.status statename,
			"RequestName" reqname,
			-- lps."TestSystemCallType",
			lps."StatusName" testtype,
			"Destination" destname,
			"Supplier" supplier,
			lpur."Nachname" as username,
			"Request_Date_Time" req_dt,
			"Tested_From" start_dt,
			"Tested_Until" end_dt,
			"Test_ASR" asr,
			"Test_ACD" acd,
			tr.succ suc_calls,
			COALESCE(tr.failed, "Test_Calls") failed_calls,
			"Test_Minutes" minutes,
			"Test_Comment" reqcomment,
			"Test_Result" reqresult,
			"TestingSystemRequestID" callingsys_id
		FROM mtcarrierdbret."Purch_Oppt" r 
		JOIN web_backend__routing.v_callingsys_request_states_lp req ON req.id	=	CASE 
																						WHEN "RequestState" = 1 THEN 1
																						WHEN ("RequestState" = 2 AND r."Tested_Until" IS NULL) THEN 2
																						WHEN ("RequestState" = 2 AND r."Tested_Until" IS NOT NULL) THEN 3
																						WHEN "RequestState" = 3 THEN 4 
																					END 
		LEFT JOIN mtcarrierdbret."Mitarb" lpur ON lpur."Personid" = r."Request_by_User"
		LEFT JOIN mtcarrierdbret."CallingSys_RouteList" lpr ON lpr."RouteID" = r."CallingSys_RouteID"
		LEFT JOIN mtcarrierdbret."Purch_Statuses" lps ON lps."StatusID" = r."Test_Type"
		-- LEFT JOIN mtcarrierdbret."CallingSys_Settings" lpcs ON lpcs."SystemID" = lps."TestSystem"
		LEFT JOIN (SELECT "CallListID" callingsysid, COUNT(*) FILTER (WHERE "CallDuration" > 0) succ, COUNT(*) FILTER (WHERE COALESCE("CallDuration", 0) = 0) failed
					FROM mtcarrierdbret."CallingSys_TestResults" 
				   	GROUP BY "CallListID") tr ON tr.callingsysid = "TestingSystemRequestID"
		ORDER BY "RequestID" DESC;
$function$
;
---------------------------------------------------------------------------------------------------------
-- Получение routeID
CREATE OR REPLACE FUNCTION web_backend__routing.f_callingsys_supplier_route_get(i_supplierid integer)
 RETURNS TABLE(routeid integer, routename character varying)
 LANGUAGE sql
AS $function$
	SELECT "RouteID", "Remote_Route_Name" 
		FROM mtcarrierdbret."CallingSys_RouteList" 
		WHERE "Captura_CarrierID" = i_supplierid
		ORDER BY 2;
$function$
;
---------------------------------------------------------------------------------------------------------
--Спсиок доступных направлений, в зависимости от тестовой системы. Для отображения в выпадающем списке Destination
CREATE OR REPLACE FUNCTION web_backend__routing.f_callingsys__lp_available_destinations(i_routingtype integer)
 RETURNS TABLE(destid integer, destname character varying)
 LANGUAGE plpgsql
AS $function$
BEGIN
	IF i_routingtype IS NULL THEN 
		RETURN QUERY
			SELECT DISTINCT "Destinationid", "Destination" 
				FROM mtcarrierdbret."Route_MasterDest_RType" 
				WHERE current_date >= "Validon" AND current_date < "Validuntil"
				ORDER BY 2;	
	ELSE
		RETURN QUERY
			SELECT "Destinationid", "Destination" 
				FROM mtcarrierdbret."Route_MasterDest_RType" 
				WHERE current_date >= "Validon" AND current_date < "Validuntil" AND "RType" = i_routingtype
				ORDER BY 2;
	END IF;
END;
$function$
;
---------------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION mtcarrierdbret.f_callingsys_sync_assure_mada_sms_trunks_get()
 RETURNS TABLE(capt_carrier character varying, capt_carrierid integer, capt_trunk character varying, assure_carrier character varying, assure_trunk character varying, assure_routeid integer)
 LANGUAGE sql
AS $function$
	-- Captura <-> Assure matching
	WITH capt_routes AS (
		-- !!have to be replaced on production with OpenMind_Suppliers table!!! ----
		SELECT regexp_replace(trunk_out, '(^V_|^CB_V_|_SMPP)', '', 'gi') trunk, trunk_out, carrier, carrier_id FROM tmp_mada_sms_supplier )	
	
	SELECT t2.carrier, t2.carrier_id, trunk_out,
			t1.name assure_carrier, t1.carrier assure_trunk, t1.sms_route_id
		FROM capt_routes t2
		LEFT JOIN  mtcarrierdbret."CallingSys_assure_sms_routes" t1 ON trunk iLike name
		--WHERE t1.name IS NULL
		ORDER BY 2;
$function$
;

-- insert into mtcarrierdbret."CallingSys_RouteList" ("CallingSystemID","Captura_CarrierID","Remote_Route_Name","Remote_Route_ID") 
-- select (3,capt_carrierid,assure_carrier,assure_routeid) from mtcarrierdbret.f_callingsys_sync_assure_mada_sms_trunks_get();
---------------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION web_backend__routing.f_callingsys__lp_sms_templates()
RETURNS TABLE(template_id integer, template_name character varying)
LANGUAGE plpgsql
AS $function$
BEGIN
	RETURN QUERY
				SELECT sms_template_id, "name" FROM mtcarrierdbret."CallingSys_assure_sms_templates";
END;
$function$
;
---------------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION web_backend__routing.f_callingsys_request_sms_list_get()
 RETURNS TABLE(reqid integer, stateid integer, statename character varying, reqname character varying, testtype character varying, destname character varying, supplier character varying, username character varying, req_dt timestamp without time zone, start_dt timestamp without time zone, end_dt timestamp without time zone, comp_calls bigint, incomp_calls bigint, reqcomment text, reqresult character varying, callingsys_id character varying)
 LANGUAGE sql
AS $function$
	SELECT
			"RequestID" reqid,
			req.id, /*"RequestState" stateid,*/
			req.status statename,
			"RequestName" reqname,
			lps."StatusName" testtype,
			"Destination" destname,
			"Supplier" supplier,
			lpur."Nachname" as username,
			"Request_Date_Time" req_dt,
			"Tested_From" start_dt,
			"Tested_Until" end_dt,
			tr.succ suc_calls,
			COALESCE(tr.failed, "Test_Calls") failed_calls,
			"Test_Comment" reqcomment,
			"Test_Result" reqresult,
			"TestingSystemRequestID" callingsys_id
		FROM mtcarrierdbret."Purch_Oppt" r 
		JOIN web_backend__routing.v_callingsys_request_states_lp req ON req.id	=	CASE 
																						WHEN "RequestState" = 1 THEN 1
																						WHEN ("RequestState" = 2 AND r."Tested_Until" IS NULL) THEN 2
																						WHEN ("RequestState" = 2 AND r."Tested_Until" IS NOT NULL) THEN 3
																						WHEN "RequestState" = 3 THEN 4 
																					END 
		LEFT JOIN mtcarrierdbret."Mitarb" lpur ON lpur."Personid" = r."Request_by_User"
		LEFT JOIN mtcarrierdbret."CallingSys_RouteList" lpr ON lpr."RouteID" = r."CallingSys_RouteID"
		LEFT JOIN mtcarrierdbret."Purch_Statuses" lps ON lps."StatusID" = r."Test_Type"
		LEFT JOIN (SELECT test_batch_id callingsysid, COUNT(*) FILTER (WHERE ui_test_status_display = 'Successful') succ, COUNT(*) FILTER (WHERE ui_test_status_display = 'Failed') failed
					FROM mtcarrierdbret."CallingSys_TestResultsAssureSMS" 
				   	GROUP BY test_batch_id) tr ON tr.callingsysid = "TestingSystemRequestID"
		ORDER BY "RequestID" DESC;
$function$
;
---------------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION web_backend__routing.f_callingsys_request_sms_get(i_requestid integer)
 RETURNS TABLE(reqname character varying, network character varying, supplier character varying, route character varying, date timestamp without time zone, senttime timestamp without time zone, result character varying, status character varying)
 LANGUAGE sql
AS $function$
	SELECT
		po."RequestName" reqname,
		r.network network,
		r.supplier supplier,
		r.route route,
		po."Request_Date_Time" date
		r.sent_time senttime,
		r.result result,
		r.ui_test_status_display status
	FROM mtcarrierdbret."CallingSys_TestResultsAssureSMS" r
		JOIN mtcarrierdbret."Purch_Oppt" po ON po."TestingSystemRequestID" = r.test_batch_id
	WHERE po."RequestID" = i_requestid
	ORDER BY r.call_batch_item_id;
$function$
;
---------------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION web_backend__routing.f_callingsys_request_state_upd(i_requestid integer, i_stateid integer, i_resultid integer, t_comment text, OUT status character varying, OUT message character varying)
 RETURNS record
 LANGUAGE plpgsql
AS $function$
DECLARE
	resname	varchar;
	s_calllistid	varchar;
	i_testingsys	integer;
	i_testype		integer;
	s_callbatch		varchar;
BEGIN
	status	:= 'FAILED';
	message	:= 'Invalid Test Parameters';
	IF i_requestid IS NULL THEN
		message	:= 'RequestID is empty!';
		RETURN;
	END IF;	
	message	:= '';
	-- Leave resname as NULL when i_resultid IS NULL. In other case get a name
	IF i_resultid IS NOT NULL THEN 
		SELECT statusname INTO resname FROM web_backend__routing.f_callingsys__lp_test_results(i_resultid);
	END IF;
	-- REMOVE TEST RESULTS
	IF i_stateid <= 2 THEN 
		SELECT "CallID", "TestSystem", "Test_Type", "TestingSystemRequestID" 
		INTO s_calllistid, i_testingsys, i_testype, s_callbatch
			FROM mtcarrierdbret."Purch_Oppt" WHERE "RequestID" = i_requestid; 
			 
		-- Delete pictures
		DELETE FROM mtcarrierdbret."CallingSys_testfiles_web"
			WHERE callid = s_calllistid AND testsystem = i_testingsys;

		-- Delete test calls
		IF i_testingsys = 3 AND i_testype IN (-67) THEN
			-- ASSURE SMS
			DELETE FROM mtcarrierdbret."CallingSys_TestResultsAssureSMS"
				WHERE test_batch_id = s_callbatch;		
		ELSE
			DELETE FROM mtcarrierdbret."CallingSys_TestResults"
				WHERE "CallListID" = s_calllistid;	
		END IF;
	END IF;
	
	UPDATE mtcarrierdbret."Purch_Oppt" SET 
			-- RESET End Timestamp WHEN NEEDED
			"TestingSystemRequestID"	= CASE WHEN i_stateid = 1 THEN NULL ELSE "TestingSystemRequestID" END,
			"Tested_Until"	= CASE WHEN i_stateid <= 2 THEN NULL ELSE COALESCE("Tested_Until", CURRENT_TIMESTAMP) END,
			"Tested_From"	= CASE WHEN i_stateid = 1 THEN NULL ELSE "Tested_From" END,  
			"RequestState"	= CASE WHEN i_stateid > 2 THEN i_stateid -1 ELSE i_stateid END,
			"Test_Result"	= resname,
			"Test_Comment"	= t_comment
		WHERE "RequestID" = i_requestid;
	GET DIAGNOSTICS message = ROW_COUNT;
		
	message	:= 'Request updated: ' || message;
	status	:= 'OK';		
END
$function$
;
---------------------------------------------------------------------------------------------------------
-- New test request
SELECT po."Request_by_User",
		po."RequestID",
		COALESCE(po."TestingSystemRequestID",'') "TestingSystemRequestID",
		po."RequestState", 
		rt."Remote_Route_ID",
		po."SupplierID",
		po."Test_Calls",
		COALESCE(po."Test_Comment",'') "Test_Comment",
		COALESCE(po."Custom_BNumbers",'') "Custom_BNumbers",
		po."Destination",
		COALESCE(dl.remote_destination_id, -1) remote_destination_id,
		COALESCE(ast.name,'') sms_template_name,
		ss."SystemID",
		ss."SystemName",
		ps."TestSystemCallType"
		FROM mtcarrierdbret."Purch_Oppt" po
		JOIN mtcarrierdbret."Purch_Statuses" ps ON po."Test_Type"=ps."StatusID"
		JOIN mtcarrierdbret."CallingSys_DestinationList" dl ON po."DestinationID"=dl.captura_destination_id AND dl.callingsys_id=ps."TestSystem"
		JOIN mtcarrierdbret."CallingSys_Settings" ss ON ss."SystemID"=ps."TestSystem"
		JOIN mtcarrierdbret."CallingSys_assure_sms_templates" ast ON ast.sms_template_id = po.sms_template_id
		LEFT JOIN mtcarrierdbret."CallingSys_RouteList" rt ON po."CallingSys_RouteID" = rt."RouteID" 
		WHERE po."Tested_Until" IS NULL 
		AND (po."TestingSystemRequestID" IS NULL OR po."TestingSystemRequestID"<>'-1') 
		AND ss."SystemName"='Assure' AND po."RequestState" < 3;
---------------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION mtcarrierdbret.f_callingsys_sync_assure_mada_sms_trunks_get(
	)
    RETURNS TABLE(capt_carrier character varying, capt_carrierid integer, capt_trunk character varying, assure_carrier character varying, assure_trunk character varying, assure_routeid integer) 
    LANGUAGE 'sql'

    COST 100
    VOLATILE 
    ROWS 1000
AS $BODY$
	-- Captura <-> Assure matching
	WITH capt_routes AS (
		-- !!have to be replaced on production with OpenMind_Suppliers table!!! ----
		SELECT regexp_replace("TrunkOut", '(^V_|^CB_V_|_SMPP)', '', 'gi') trunk, "TrunkOut", "Carrier", "CarrierID" FROM mtcarrierdbret."OpenMind_Supplier")	
	
	SELECT t2."Carrier", t2."CarrierID", "TrunkOut",
			t1.name assure_carrier, t1.carrier assure_trunk, t1.sms_route_id
		FROM capt_routes t2
		LEFT JOIN  mtcarrierdbret."CallingSys_assure_sms_routes" t1 ON trunk iLike name
		--WHERE t1.name IS NULL
		ORDER BY 2;
$BODY$;

ALTER FUNCTION mtcarrierdbret.f_callingsys_sync_assure_mada_sms_trunks_get()
    OWNER TO ocuconnection;

---------------------------------------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION web_backend__routing.f_callingsys__lp_available_destinations(
    i_routingtype integer)
    RETURNS TABLE(destid integer, destname character varying)
   LANGUAGE 'plpgsql'

   COST 100
   VOLATILE
   ROWS 1000
AS $BODY$
BEGIN
    IF i_routingtype IS NULL THEN
        RETURN QUERY
            SELECT DISTINCT t."Destinationid", t."Destination"
                FROM mtcarrierdbret."Route_MasterDest_RType" t
				JOIN mtcarrierdbret."CallingSys_DestinationList" t2 ON t."Destinationid" = t2.captura_destination_id
                WHERE current_date >= t."Validon" AND current_date < t."Validuntil" AND t2.remote_destination_id NOTNULL
                ORDER BY 2;    
    ELSE
        RETURN QUERY
            SELECT t."Destinationid", t."Destination"
                FROM mtcarrierdbret."Route_MasterDest_RType" t
				JOIN mtcarrierdbret."CallingSys_DestinationList" t2 ON t."Destinationid" = t2.captura_destination_id
                WHERE current_date >= t."Validon" AND current_date < t."Validuntil" AND t."RType" = i_routingtype AND t2.remote_destination_id NOTNULL
                ORDER BY 2;
    END IF;
END;
$BODY$;

ALTER FUNCTION web_backend__routing.f_callingsys__lp_available_destinations(integer)
   OWNER TO web_backend;