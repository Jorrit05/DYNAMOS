class DynamosApi
  def stimulate_dynamos(endpoint = 'requestApproval', url = 'http://api-gateway.api-gateway.svc.cluster.local:8080/api/v1')
    uri = "#{url}/#{endpoint}"
    request_properties = { 'Content-Type' => 'application/json' }
    http = Net::HTTP.new(uri.host, uri.port)
    Net::HTTP::Post.new(uri.request_uri, request_properties)

    body = request_body
    logger.info("Stimulating SUT with body: #{body}")
    response = http.request(body)
    logger.info("Response from SUT: #{response.body}")

    response
  end

  def request_body
    {
      type: 'sqlDataRequest',
      user: {
        id: '12324',
        userName: 'jorrit.stutterheim@cloudnation.nl'
      },
      dataProviders: %w[VU UVA RUG],
      data_request: {
        type: 'sqlDataRequest',
        query: 'SELECT * FROM Personen p JOIN Aanstellingen s LIMIT 1000',
        # // "query" : "SELECT p.Geslacht, s.Salschal FROM Personen p JOIN Aanstellingen s ON p.Unieknr = s.Unieknr",
        # // "query" : "SELECT DISTINCT p.Unieknr, p.Geslacht, p.Gebdat, s.Aanst_22, s.Functcat, s.Salschal as Salary FROM Personen p JOIN Aanstellingen s ON p.Unieknr = s.Unieknr LIMIT 4",
        algorithm: 'average',
        # // "algorithmColumns" : {
        # //     "Geslacht" : "Aanst_22, Gebdat"
        # // },
        options: {
          graph: false,
          aggregate: false
        },
        requestMetadata: {}
      }
    }.to_json
  end
end
