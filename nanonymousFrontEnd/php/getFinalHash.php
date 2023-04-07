<?php
header('Content-Type: text/event-stream');
// recommended to prevent caching of event data.
header('Cache-Control: no-cache');
// Needed for nginx servers
header('X-Accel-Buffering: no');

function info($response) {
   echo $response;

   ob_flush();
   flush();
}

$finalAddress=$_GET['address'];

$context = stream_context_create(
   ['ssl' => [
     'allow_self_signed' => true
]]);

$socket = stream_socket_client('ssl://147.182.231.89:41721', $errno, $errstr, 30, STREAM_CLIENT_CONNECT, $context);
if (!$socket) {
   echo "comm error: $errstr ($errno)<br />\n";
} else {
   fwrite($socket, "trRequest&address=$finalAddress=");
   while (true) {

      $buffer = fgets($socket);
      if ($buffer !== false) {
         if (str_contains($buffer, "hash")) {
            $hash = $buffer;
            break;
         } else {
            info($buffer);
         }
      }
   }
   fclose($socket);
}

echo $hash;
?>
