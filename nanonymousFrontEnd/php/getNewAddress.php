<?php
$finalAddress=$_GET['address'];

$context = stream_context_create(
   ['ssl' => [
     'allow_self_signed' => true
]]);

$socket = stream_socket_client('ssl://147.182.231.89:41721', $errno, $errstr, 30, STREAM_CLIENT_CONNECT, $context);
if (!$socket) {
   echo "comm error: $errstr ($errno)<br />\n";
} else {
   fwrite($socket, "newaddress&address=$finalAddress=");
   while (($buffer = fgets($socket, 128)) !== false) {
      $newAddress = $buffer;
   }
   fclose($socket);
}

echo $newAddress;
?>
