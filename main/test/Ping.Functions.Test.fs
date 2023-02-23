module PingFunctionsTests

// testing libraries
open Expecto
open FsCheck
open JustEat.HttpClientInterception // mocking library for HttpClient

// system libraries
open System.Net.Http

open Domain.Environment

// module under test
open Domain.Ping
open Functions.Ping

let options =
    HttpClientInterceptorOptions()
        .ThrowsOnMissingRegistration()

let builder =
    HttpRequestInterceptionBuilder()
        .Requests()
        .ForHttps()
        .ForPost()
        .ForHost("api-ipv4.porkbun.com")
        .ForPath("api/json/v3/ping")
        .Responds()
        .WithContent(
            { PBPingSuccessResponse.Status = "Success"
              PBPingSuccessResponse.YourIP = "192.168.1.1" }
            |> PBPingSuccessResponse.encoder
            |> (fun x -> x.ToString())
        )
        .RegisterWith(options)

[<Tests>]
let fetchIPTests =
    testList
        "fetchIP tests"

        [ testCase "Runs the test with the interceptor, returning the IP Address"
          <| fun _ ->
              let expected = "192.168.1.1" |> IPAddress |> Ok

              use client = options.CreateHttpClient()

              let cmd =
                  { PBPingCommand.APIKey = APIKey "pk1_...."
                    PBPingCommand.SecretKey = SecretKey "sk1_...." }

              async {
                  let! result = fetchIP client cmd

                  Expect.equal result expected "did not return a success with the correct IP address"
              }
              |> Async.RunSynchronously

         ]