<?php
// Same as getNewAddress.php but returns standard JSON

$request = explode("?", $_SERVER['REQUEST_URI']);
$context = stream_context_create(
   ['ssl' => [
     'allow_self_signed' => true
]]);

// TODO real server
//$socket = stream_socket_client('ssl://147.182.231.89:41721', $errno, $errstr, 30, STREAM_CLIENT_CONNECT, $context);
$socket = stream_socket_client('ssl://164.92.92.51:41721', $errno, $errstr, 30, STREAM_CLIENT_CONNECT, $context);
if (!$socket) {
   echo "{\"error\": \"comm error: $errstr ($errno)\n\"}";
} else {
   if (count($request) > 1) {
      fwrite($socket, $request[1] .'&api');
      while (($buffer = fgets($socket, 128)) !== false) {
         $newAddress = $buffer;
      }
      fclose($socket);
   } else {
      echo "{\"error\": \"Malformed URI\n\"}";
   }
}

$json = "{";
// Construct the JSON
$parameters = explode("&", $newAddress);
for ($i = 0; $i < count($parameters); $i++) {
   $statement = explode("=", $parameters[$i]);
   if (count($statement) > 1) {
      if ($i != 0) {
         $json .= ', ';
      }
      $json .= '"'. $statement[0] .'": ';
      $values = explode(",", $statement[1]);
      if (count($values) > 1) {
         $json .= "[";
         for ($j = 0; $j < count($values); $j++) {
            if ($j != 0) {
               $json .= ', ';
            }
            if (is_numeric($values[$j])) {
               $json .= $values[$j];
            } else {
               $json .= '"'. $values[$j] .'"';
            }
         }
         $json .= "]";
      } else {
         if (is_numeric($values[0])) {
            $json .= $values[0];
         } else {
            $json .= '"'. $values[0] .'"';
         }
      }
   }
}
$json .= "}";

echo $json;
?>
