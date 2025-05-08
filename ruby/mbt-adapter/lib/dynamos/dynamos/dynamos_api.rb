class DynamosApi
  def initialize(api_gateway_url = 'http://api-gateway.api-gateway.svc.cluster.local:8080/api/v1')
    @api_gateway_url = api_gateway_url
  end

  def stimulate_dynamos(endpoint = 'requestApproval')
    uri = URI.parse("#{@api_gateway_url}/#{endpoint}")
    request_properties = { 'Content-Type' => 'application/json' }

    logger.info("Stimulating SUT with body: #{request_body}")
    response = Net::HTTP.post(uri, request_body, request_properties)
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
