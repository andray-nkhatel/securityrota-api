#!/bin/bash
# Import historical shifts from previous manual schedule

BASE_URL="http://localhost:8080/api/v1"

echo "Importing historical shifts for week of 2025-11-23..."

curl -s -X POST "$BASE_URL/admin/import-shifts" \
  -H "Content-Type: application/json" \
  -d '{
  "shifts": [
    {"name":"Sgt. Kalongana","date":"2025-11-23","shift_type":"day","status":"on_duty"},
    {"name":"Sgt. Kalongana","date":"2025-11-24","shift_type":"day","status":"on_duty"},
    {"name":"Sgt. Kalongana","date":"2025-11-25","shift_type":"day","status":"on_duty"},
    {"name":"Sgt. Kalongana","date":"2025-11-26","shift_type":"day","status":"on_duty"},
    {"name":"Sgt. Kalongana","date":"2025-11-27","shift_type":"day","status":"on_duty"},
    {"name":"Sgt. Kalongana","date":"2025-11-28","shift_type":"day","status":"on_duty"},
    {"name":"Sgt. Kalongana","date":"2025-11-29","shift_type":"day","status":"off_duty"},

    {"name":"Faides","date":"2025-11-23","shift_type":"day","status":"off_duty"},
    {"name":"Faides","date":"2025-11-24","shift_type":"day","status":"on_duty"},
    {"name":"Faides","date":"2025-11-25","shift_type":"day","status":"on_duty"},
    {"name":"Faides","date":"2025-11-26","shift_type":"day","status":"on_duty"},
    {"name":"Faides","date":"2025-11-27","shift_type":"day","status":"on_duty"},
    {"name":"Faides","date":"2025-11-28","shift_type":"day","status":"on_duty"},
    {"name":"Faides","date":"2025-11-29","shift_type":"day","status":"on_duty"},

    {"name":"Abigail","date":"2025-11-23","shift_type":"day","status":"on_duty"},
    {"name":"Abigail","date":"2025-11-24","shift_type":"day","status":"on_duty"},
    {"name":"Abigail","date":"2025-11-25","shift_type":"day","status":"on_duty"},
    {"name":"Abigail","date":"2025-11-26","shift_type":"day","status":"on_duty"},
    {"name":"Abigail","date":"2025-11-27","shift_type":"day","status":"on_duty"},
    {"name":"Abigail","date":"2025-11-28","shift_type":"day","status":"on_duty"},
    {"name":"Abigail","date":"2025-11-29","shift_type":"day","status":"off_duty"},

    {"name":"Alexander","date":"2025-11-23","shift_type":"day","status":"on_duty"},
    {"name":"Levy","date":"2025-11-23","shift_type":"day","status":"on_duty"},
    {"name":"Anderson","date":"2025-11-23","shift_type":"day","status":"on_duty"},
    {"name":"Julius","date":"2025-11-23","shift_type":"day","status":"on_duty"},
    {"name":"Richard","date":"2025-11-23","shift_type":"day","status":"on_duty"},
    {"name":"Rodger","date":"2025-11-23","shift_type":"day","status":"on_duty"},
    {"name":"Chipoya","date":"2025-11-23","shift_type":"day","status":"on_duty"},
    {"name":"Mubiana","date":"2025-11-23","shift_type":"day","status":"on_duty"},

    {"name":"Moses","date":"2025-11-23","shift_type":"night","status":"on_duty"},
    {"name":"Michael","date":"2025-11-23","shift_type":"night","status":"on_duty"},
    {"name":"Gift","date":"2025-11-23","shift_type":"night","status":"on_duty"},
    {"name":"Gibson","date":"2025-11-23","shift_type":"night","status":"on_duty"},
    {"name":"Sinoya","date":"2025-11-23","shift_type":"night","status":"on_duty"},
    {"name":"Hendrix","date":"2025-11-23","shift_type":"night","status":"on_duty"},
    {"name":"Arthur","date":"2025-11-23","shift_type":"night","status":"on_duty"},
    {"name":"Paul","date":"2025-11-23","shift_type":"night","status":"on_duty"}
  ]
}' | jq .

echo ""
echo "Setting rotation state for that week..."
curl -s -X POST "$BASE_URL/admin/import-state" \
  -H "Content-Type: application/json" \
  -d '{"week_start":"2025-11-23","day_shift_team":1}' | jq .

echo ""
echo "Historical shifts imported. Query them:"
echo "curl '$BASE_URL/shifts?week_start=2025-11-23'"
