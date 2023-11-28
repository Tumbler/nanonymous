<?php
session_start();

if (isset($_SESSION[$_SERVER['REMOTE_ADDR']])) {
   $timeElapsed = time() - (int)$_SESSION[$_SERVER['REMOTE_ADDR']];
   if ($timeElapsed < 30) {
      //echo "info: Please wait 30 seconds before requesting another address";
      die("info=Please wait 30 seconds before requesting another address.\n");
   } else {
      $_SESSION[$_SERVER['REMOTE_ADDR']] = time();
   }
} else {
   $_SESSION[$_SERVER['REMOTE_ADDR']] = time();
}

$request = explode("?", $_SERVER['REQUEST_URI']);
$context = stream_context_create(
   ['ssl' => [
     'allow_self_signed' => true
]]);

$socket = stream_socket_client('ssl://147.182.231.89:41721', $errno, $errstr, 30, STREAM_CLIENT_CONNECT, $context);
if (!$socket) {
   echo "comm error: $errstr ($errno)<br />\n";
} else {
   if (count($request) > 1) {
      fwrite($socket, $request[1]);
      while (($buffer = fgets($socket, 128)) !== false) {
         $newAddress = $buffer;
      }
      fclose($socket);
   } else {
      echo "error: Malformed URI<br />\n";
   }
}

echo $newAddress;
?>
