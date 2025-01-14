# frozen_string_literal: true

require 'twilio-ruby'
require 'sinatra'
require 'sinatra/json'
require 'dotenv'
require 'faker'
require 'pry'
require 'nokogiri'
# Load environment configuration
Dotenv.load

def twilio_number
  ENV['TWILIO_CALLER_ID']
end

# Create a random username for the client
IDENTITY = Array.new(1)

# Render home page
get '/' do
  File.read(File.join('public', 'index.html'))
end

# Generate a token for use in our Video application
get '/token' do
  # Required for any Twilio Access Token
  account_sid = ENV['TWILIO_ACCOUNT_SID']
  api_key = ENV['API_KEY']
  api_secret = ENV['API_SECRET']

  # Required for Voice
  outgoing_application_sid = ENV['TWILIO_TWIML_APP_SID']


  # Create Voice grant for our token
  grant = Twilio::JWT::AccessToken::VoiceGrant.new
  grant.outgoing_application_sid = outgoing_application_sid

  # Optional: add to allow incoming calls
  grant.incoming_allow = true

  identity = Faker::Internet.user_name.gsub(/[^0-9a-z_]/i, '')
  IDENTITY.clear
  IDENTITY.append(identity)

  # Create an Access Token
  token = Twilio::JWT::AccessToken.new(
    account_sid,
    api_key,
    api_secret,
    [grant],
    identity: IDENTITY[0]
  )

  # Generate the token and send to client
  json identity: IDENTITY[0], token: token.to_jwt
end

post '/conference' do
  puts "JOJNT: #{params}"
  conference = params['jojnt_conference']

  # https://www.twilio.com/docs/voice/twiml/conference#
  response = Twilio::TwiML::VoiceResponse.new
  response.dial do |dial|
    dial.conference(
      conference,
      wait_url: '', # empty for no music
      # beep: false,
      start_conference_on_enter: true,
      end_conference_on_exit: true,
    )
  end

  puts "TWIML"
  puts response
  content_type 'text/xml'
  response.to_s
end

post '/status' do
  puts "Twilio called back"
  puts "PARAMS #{params}"
end


post '/voice' do
  twiml = Twilio::TwiML::VoiceResponse.new do |r|
    if params['To'] && params['To'] == twilio_number
      r.dial do |d|
        d.client(identity: IDENTITY[0])
      end
    elsif params['To'] && params['To'] != ''
      r.dial(caller_id: twilio_number) do |d|
        # wrap the phone number or client name in the appropriate TwiML verb
        # by checking if the number given has only digits and format symbols
        if params['To'] =~ /^[\d+\-() ]+$/
          d.number(params['To'])
        else
          d.client identity: params['To']
        end
      end
    else
      r.say(message: 'Thanks for calling!')
    end
  end

  puts twiml
  content_type 'text/xml'
  twiml.to_s
end
