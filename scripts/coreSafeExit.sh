# Doing it twice is intentional and needed
curl -s -o /dev.null -k --tlsv1.2 --request POST --data "safeExit" https://localhost:41721
curl -s -o /dev.null -k --tlsv1.2 --request POST --data "safeExit" https://localhost:41721
