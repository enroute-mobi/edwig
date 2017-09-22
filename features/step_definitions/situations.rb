def situations_path(attributes = {})
  url_for_model(attributes.merge(resource: 'situation'))
end

def situation_path(id, attributes = {})
  path = url_for_model(attributes.merge(resource: 'situation', id: id))
end



Given(/^a Situation exists (?:in Referential "([^"]+)" )?with the following attributes:$/) do |referential, situation|
  RestClient.post situations_path(referential: referential), model_attributes(situation).to_json, {content_type: :json, :Authorization => "Token token=#{$token}" }
end

When(/^a Situation is created (?:in Referential "([^"]+)" )?with the following attributes:$/) do |referential, situation|
  if referential.nil?
    step "a Situation exists with the following attributes:", situation
  else
    step "a Situation exists in Referential \"#{referential}\" with the following attributes:", situation
  end
end

When(/^the Situation "([^"]+)":"([^"]+)"(?: in Referential "([^"]+)")? is destroyed$/) do |kind, objectid, referential|
  response = RestClient.get situations_path(referential: referential), {content_type: :json, :Authorization => "Token token=#{$token}"}
  responseArray = JSON.parse(response.body)
  expectedSituation = responseArray.find{|a| a["ObjectIDs"][kind] == objectid }

  RestClient.delete situation_path(expectedSituation["Id"]), {:Authorization => "Token token=#{$token}"}
end


Then(/^one Situation(?: in Referential "([^"]+)")? has the following attributes:$/) do |referential, attributes|
  response = RestClient.get situations_path(referential: referential), {content_type: :json, :Authorization => "Token token=#{$token}" }
  response_array = JSON.parse(response.body)

  called_method = has_attributes(response_array, attributes)

  expect(called_method).to be_truthy
end

Then(/^a Situation "([^"]+)":"([^"]+)" should( not)? exist(?: in Referential "([^"]+)")?$/) do |kind, value, condition, referential|
  response = RestClient.get(situation_path("#{kind}:#{value}" ,referential: referential), {content_type: :json, :Authorization => "Token token=#{$token}"} ){|response, request, result| response }

  if condition.nil?
    expect(response.code).to eq(200)
  else
    expect(response.code).to eq(404)
    expect(response.body).to include("Situation not found: #{kind}:#{value}")
  end
end

Then(/^the Situation "([^"]*)" has the following attributes:$/) do |identifier, attributes|
  # For tests
  # puts RestClient.get situations_path, {Authorization: "Token token=#{$token}"}

  response = RestClient.get situation_path(identifier), {content_type: :json, :Authorization => "Token token=#{$token}"}
  situationAttributes = api_attributes(response.body)
  expect(situationAttributes).to include(model_attributes(attributes))
end
