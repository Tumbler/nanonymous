<?php
session_start();

$newCurrency=$_GET['curr'];

// Simple cache so we don't have to hit the file every time.
if (isset($_SESSION['CURRENCY_JSON'])) {
   $json_data = $_SESSION['CURRENCY_JSON'];
} else {
   // Load it from file
   $json = file_get_contents('./exchanges.json');

   // Decode the JSON file
   $json_data = json_decode($json,true);

   // Cache for later
   $_SESSION['CURRENCY_JSON'] = $json_data;
}

// Give the rate requested.
echo "val=". $json_data['rates']['USD'] .",". $json_data['rates'][$newCurrency];

?>
