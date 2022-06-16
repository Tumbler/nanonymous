<?php
$finalAddress=$_GET['address'];

$socket = fsockopen("147.182.231.89", 41721, $errno, $errstr, 30);
if (!$socket) {
   echo "comm error: $errstr ($errno)<br />\n";
} else {
   fwrite($socket, "trRequest&address=$finalAddress=");
   while (($buffer = fgets($socket, 128)) !== false) {
      $hash = $buffer;
   }
   fclose($socket);
}

echo $hash;
?>
