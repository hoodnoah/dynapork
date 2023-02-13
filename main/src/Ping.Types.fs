namespace Domain

open System.Net.Http

open Thoth.Json.Net

module Ping =
    type PBPingCommand =
        { SecretAPIKey: string
          APIKey: string }

    type PBPingSuccessResponse = { Status: string; YourIP: string }

    type PBPingFailureResponse = { Status: string; Message: string }

    type IPAddress = IPAddress of string

    type PBPingResponse =
        | PBPingSuccess of PBPingSuccessResponse
        | PBPingFailure of PBPingFailureResponse

    type PBPingError =
        | InvalidAPIKey
        | JSONDecodeFailure of string
        | Unknown

    type PBPingResult = Result<IPAddress, PBPingError>

    type FetchIP = HttpClient -> PBPingCommand -> Async<PBPingResult>

    // Thoth coders
    module PBPingCommand =
        let encoder (cmd: PBPingCommand) =
            Encode.object [ "secretapikey", Encode.string cmd.SecretAPIKey
                            "apikey", Encode.string cmd.APIKey ]

        let decoder: Decoder<PBPingCommand> =
            Decode.object (fun get ->
                { PBPingCommand.SecretAPIKey = get.Required.Field "secretapikey" Decode.string
                  PBPingCommand.APIKey = get.Required.Field "apikey" Decode.string })

    module PBPingSuccessResponse =
        let encoder (successResponse: PBPingSuccessResponse) =
            Encode.object [ "status", Encode.string successResponse.Status
                            "yourIp", Encode.string successResponse.YourIP ]

        let decoder: Decoder<PBPingSuccessResponse> =
            Decode.object (fun get ->
                { PBPingSuccessResponse.Status = get.Required.Field "status" Decode.string
                  PBPingSuccessResponse.YourIP = get.Required.Field "yourIp" Decode.string })

    module PBPingFailureResponse =
        let encoder (failureResponse: PBPingFailureResponse) =
            Encode.object [ "status", Encode.string failureResponse.Status
                            "message", Encode.string failureResponse.Message ]

        let decoder: Decoder<PBPingFailureResponse> =
            Decode.object (fun get ->
                { PBPingFailureResponse.Status = get.Required.Field "status" Decode.string
                  PBPingFailureResponse.Message = get.Required.Field "message" Decode.string })

    module PBPingResponse =
        let encoder (resp: PBPingResponse) =
            match resp with
            | PBPingSuccess s ->
                Encode.object [ "status", Encode.string s.Status
                                "yourIp", Encode.string s.YourIP ]
            | PBPingFailure f ->
                Encode.object [ "status", Encode.string f.Status
                                "message", Encode.string f.Message ]

        let decoder: Decoder<PBPingResponse> =
            Decode.oneOf [ PBPingSuccessResponse.decoder
                           |> Decode.map PBPingSuccess
                           PBPingFailureResponse.decoder
                           |> Decode.map PBPingFailure ]