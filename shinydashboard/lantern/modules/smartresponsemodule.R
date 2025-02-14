# SMART-on-FHIR Well-known URI responses

smartresponsemodule_UI <- function(id) {

  ns <- NS(id)

  tagList(
    p("This is the SMART-on-FHIR Core Capabilities response page. FHIR endpoints
      requiring authorization shall provide a JSON document at the endpoint URL with ",
      code("/.well-known/smart-configuration"), "appended to the end of the base URL."
    ),
    fluidRow(
      column(width = 6,
             tableOutput(ns("well_known_summary_table")),
             tableOutput(ns("smart_vendor_table"))
      ),
      column(width = 6,
             tableOutput(ns("smart_capability_count_table")))
    ),
    h3("Endpoints by Well Known URI support"),
    p("This is the list of endpoints which have returned a valid SMART Core Capabilities JSON document at the", code("/.well-known/smart-configuration"), " URI."),
    DT::dataTableOutput(ns("well_known_endpoints"))
  )
}

smartresponsemodule <- function(
  input,
  output,
  session,
  sel_fhir_version,
  sel_vendor
) {

  ns <- session$ns

  # IN PROGRESS - need to get the correct query with smart_http_response and smart_response columns
  # can we show a summary table of how many endpoints supporting /.well-known/smart-configuration ?

  get_filtered_data <- function(table_val) {
  res <- table_val
  req(sel_fhir_version(), sel_vendor())
  if (sel_fhir_version() != ui_special_values$ALL_FHIR_VERSIONS) {
    res <- res %>% filter(fhir_version == sel_fhir_version())
  }
  if (sel_vendor() != ui_special_values$ALL_DEVELOPERS) {
    res <- res %>% filter(vendor_name == sel_vendor())
  }
  res
}

  selected_smart_capabilities <- reactive({
    res <- isolate(app_data$smart_response_capabilities())
    res <- get_filtered_data(res)
    res
  })

  selected_smart_count_total <- reactive({
    all <- endpoint_export_tbl
    all <- get_filtered_data(all)
    all <- all %>% distinct(url) %>% count() %>% pull(n)
    all
  })

  selected_smart_count_200 <- reactive({
    res <- isolate(app_data$http_pct())
    res <- get_filtered_data(res)
    res <- res %>%
      select(http_response) %>%
      group_by(http_response) %>%
      filter(http_response == 200) %>%
      tally()

    max((res %>% filter(http_response == 200)) %>% pull(n), 0)
  })

  selected_well_known_count_doc <- reactive({
    res <- app_data$well_known_endpoints_tbl()
    res <- get_filtered_data(res)
    res
  })

  selected_well_known_count_no_doc <- reactive({
    res <- app_data$well_known_endpoints_no_doc()
    res <- get_filtered_data(res)
    res
  })

  selected_well_known_endpoints_count <- reactive({
    res <- endpoint_export_tbl
      res <- get_filtered_data(res)
    res <- res %>% filter(smart_http_response == 200)
    res <- res %>% distinct(url) %>% count() %>% pull(n)
    res
  })

  selected_well_known_endpoint_counts <- reactive({
    res <- tribble(
      ~Status, ~Endpoints,
      "Total Indexed Endpoints", as.integer(selected_smart_count_total()),
      "Endpoints with successful response (HTTP 200)", as.integer(selected_smart_count_200()),
      "Well Known URI Endpoints with successful response (HTTP 200)", as.integer(selected_well_known_endpoints_count()),
      "Well Known URI Endpoints with valid response JSON document", as.integer(nrow(selected_well_known_count_doc())),
      "Well Known URI Endpoints without valid response JSON document", as.integer(nrow(selected_well_known_count_no_doc()))
    )
  })

  output$smart_capability_count_table <- renderTable(
    get_smart_response_capability_count(selected_smart_capabilities())
  )

  output$smart_vendor_table <- renderTable({
    selected_smart_capabilities() %>%
      distinct(id, .keep_all = TRUE) %>%
      group_by(id, fhir_version, vendor_name) %>%
      count() %>%
      group_by(fhir_version, vendor_name) %>%
      count(wt = n) %>%
      select("FHIR Version" = fhir_version, "Developer" = vendor_name, "Endpoints" = n)
  })

  output$well_known_summary_table <- renderTable(
    selected_well_known_endpoint_counts()
  )

  selected_endpoints <- reactive({
    res <- isolate(app_data$well_known_endpoints_tbl())
    res <- get_filtered_data(res)
    res
  })

  output$well_known_endpoints <-  DT::renderDataTable({
    datatable(selected_endpoints(),
              colnames = c("URL", "Organization", "Developer", "FHIR Version"),
              rownames = FALSE,
              options = list(scrollX = TRUE)
    )
  })

}
