package main

import (
   "fmt"
   "time"
   "os"
   "bufio"
   "strings"
   "regexp"
   "strconv"
)

const LOG_DIR = "./logs/"

var dateRegex, _ = regexp.Compile(`\d+/\d+/\d+ \d+:\d+`)

func lastWeekSummary() error {

   // 24 x 7 = 168
   lastWeek := time.Now().Add(-168 * time.Hour)
   lastWeek.Truncate(24 * time.Hour)

   gross, err := getProfitSince(lastWeek)
   if (err != nil) {
      Warning.Println("lastWeekSummary: ", err)
      return fmt.Errorf("lastWeekSummary: %w", err)
   }

   USD, err := getUSDSinceAtTimeOfTransaction(lastWeek)
   if (err != nil) {
      Warning.Println("lastWeekSummary: ", err)
      return fmt.Errorf("lastWeekSummary: %w", err)
   }

   numOfTransactions, err := getNumOfTransactionsSince(lastWeek)
   if (err != nil) {
      Warning.Println("lastWeekSummary: ", err)
      return fmt.Errorf("lastWeekSummary: %w", err)
   }

   transactionID, err := peekAtNextTransactionId()
   if (err != nil) {
      Warning.Println("lastWeekSummary: ", err)
      return fmt.Errorf("lastWeekSummary: %w", err)
   }

   nanoValue, err := getNanoUSDValue()
   if (err != nil) {
      Warning.Println("lastWeekSummary: ", err)
      return fmt.Errorf("lastWeekSummary: %w", err)
   }

   USDnow := rawToNANO(gross) * nanoValue

   logFiles, err := os.ReadDir(LOG_DIR)
   if (err != nil) {
      Warning.Println("lastWeekSummary: ", err)
      return fmt.Errorf("lastWeekSummary: %w", err)
   }

   logAnomalies := make(map[string][][]string)

   // Find all anomalies that happened in the last week.
   for _, dirEntry := range logFiles {
      if (dirEntry.IsDir()) {
         continue
      }

      filename := dirEntry.Name()
      logAnomalies[filename] = append(logAnomalies[filename], make([]string, 0))
      logAnomalies[filename] = append(logAnomalies[filename], make([]string, 0))

      file, err := os.Open(LOG_DIR + filename)
      if (err != nil) {
         Warning.Println("lastWeekSummary: ", err)
         return fmt.Errorf("lastWeekSummary: %w", err)
      }

      scanner := bufio.NewScanner(file)
      // Scan one line at a time.
      scanner.Split(bufio.ScanLines)

      for (scanner.Scan()) {
         line := scanner.Text()

         dateString := dateRegex.FindString(line)
         date, err := time.Parse("2006/01/02 15:04", dateString)

         if (err != nil || lastWeek.After(date)) {
            continue
         }

         if (strings.Contains(line, "WARNING:")) {
            logAnomalies[filename][0] = append(logAnomalies[filename][0], line)
         } else if (strings.Contains(line, "ERROR:")){
            logAnomalies[filename][1] = append(logAnomalies[filename][1], line)
         }
      }

   }

   var reportString string

   for file, printouts := range logAnomalies {
      if (len(printouts[0]) + len(printouts[1]) > 0) {
         reportString += file +":\n"
         reportString += "   "+ strconv.Itoa(len(printouts[0])) +" Warnings\n"
         reportString += "   "+ strconv.Itoa(len(printouts[1])) +" Errors\n"

         for _, warning := range printouts[0] {
            reportString += warning + "\n"
         }

         if (len(printouts[0]) > 0) {
            reportString += "\n\n\n"
         }

         for _, errors := range printouts[1] {
            reportString += errors + "\n"
         }

         if (len(printouts[1]) > 0) {
            reportString += "\n\n\n"
         }
      }
   }

   totalReport := fmt.Sprint(
      "\nWeek summary:",
    "\n\nGross profit: ",
      "\n   Ӿ ", rawToNANO(gross),
      "\n   $ ", USD, " (at time of transaction)",
      "\n   $ ", USDnow, " (gross vs value now)",
    "\n\nSuccessful transactions: ", numOfTransactions,
      "\nCurrent transaction ID: ", transactionID,
    "\n\nError reports:\n", reportString,
   )

   fmt.Print(totalReport)
   sendEmail("Weely Nanonymous Sumamry", totalReport)

   return nil
}
