library(dygraphs)
library(xts)

performance_UI <- function(id) {

  ns <- NS(id)

  tagList(
    textOutput(ns("no_graph")),
    dygraphOutput(ns("mean_response_time_plot")),
    p("Click and drag on plot to zoom in, double-click to zoom out."),
    htmlOutput(ns("note_text"))
  )
}

performancemodule <- function(
    input,
    output,
    session,
    sel_date
) {
  ns <- session$ns

  get_range <- function() {
    if (all(sel_date() == "Past 7 days")) {
      range <- "604800"
    }
    else if (all(sel_date() == "Past 14 days")) {
      range <- "1209600"
    }
    else if (all(sel_date() == "Past 30 days")) {
      range <- "2592000"
    }
    else{
      range <- "maxdate.maximum"
    }
    range
  }

  response_time_xts <- reactive({
    range <- get_range()
    res <- get_avg_response_time(db_connection, range)
    # convert to xts format for use in dygraph
    xts(x = cbind(res$max, res$avg, res$min),
        order.by = res$date
    )
  })

  output$no_graph <- renderText({
    if (nrow(response_time_xts()) == 0) {
      "Sorry, there isn't enough data to show response times!"
    }
  })

  output$mean_response_time_plot <- renderDygraph({
    if (nrow(response_time_xts()) > 0) {
      dygraph(response_time_xts(),
            main = "Endpoint Mean Response Time",
            ylab = "seconds",
            xlab = "Date") %>%
      dyAxis("y", valueRange = c(-1.30, NA)) %>%
      dySeries("V1", label = "Maximum") %>%
      dySeries("V2", label = "Average") %>%
      dySeries("V3", label = "Minimum") %>%
      dyLegend(width = 450)
    }
  })

  output$note_text <- renderUI({
    note_info <- "There are many variables that influence response time, such
      as network congestion, geographic location, hosting configurations, etc.
      This graphic only intends to convey the health of the FHIR endpoint ecosystem
      as a whole, drastic changes to which may represent some widespread issue
      throughout the ecosystem."
    res <- paste("<div style='font-size: 18px;'><b>Note:</b>", note_info, "</div>")
    HTML(res)
  })

}
