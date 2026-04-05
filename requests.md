# when joining an empty call:

what happens in sequential order (uninteresting bits left out like all the messenger stuff)

## initial connection

```http request
GET
    https://cloud-api.yandex.ru
    /telemost_front/v2/telemost/conferences
    /https%3A%2F%2Ftelemost.yandex.ru%2Fj%2F69080743110758
    /connection?next_gen_media_platform_allowed=true
    &display_name=cool_username&waiting_room_supported=true
```

```json5
{
  "connection_type": "CONFERENCE",
  "uri": "https://telemost.yandex.ru/j/69080743110758",
  "room_id": "d48791a2-07d5-4d3b-88d9-d82557bd56ac",
  "safe_room_id": "d48791a2-07d5-4d3b-88d9-d82557bd56ac",
  "peer_id": "f72c8dc2-6ac6-43de-819a-06e3c3f0227c",
  "client_configuration": {
    "cloud_recording_available": false,
    "wait_time_to_reconnect_ms": 4404,
    "media_server_url": "wss://goloom.strm.yandex.net/join",
    "service_name": "telemost",
    "alice_pro_enabled": true,
    "summarization_available": false,
    "new_grid_mobile": false,
    "extended_role_permissions_enabled": false,
    "calendar_summarization_receive_available": false,
    "reactions_available": true,
    "join_url_hidden": false,
    "state_check_interval_seconds": 30,
    "goloom_session_open_ms": 120000,
    "new_grid": true,
    "waiting_room_peers_refresh_interval_ms": 10000,
    "report_problem_button_available": false,
    "ice_servers": [
      {
        "urls": [
          "stun:stun.rtc.yandex.net:3478"
        ]
      }
    ]
  },
  "conference_state": {
    "local_recording_allowed": true,
    "cloud_recording_allowed": false,
    "chat_allowed": true,
    "control_allowed": true,
    "broadcast_allowed": false,
    "broadcast_feature_enabled": false,
    "access_restriction_organization_allowed": false,
    "chat_path": "0/22/577ba3bb-9b23-4bd4-b148-db18916582eb",
    "access_level": "PUBLIC",
    "summarization_status": "TURNED_OFF",
    "cloud_recording_status": "TURNED_OFF"
  },
  "media_platform": "GOLOOM",
  "is_legal_entity": false,
  "session_id": "4b2c4d25-5fb8-4710-9912-2ddd05332c15",
  "waiting_room_available": false,
  "expiration_time": 1771509742987,
  "conference_limit": 40,
  "peer_session_id": "f507fcce-5616-41f1-9693-e6800ea22d92",
  "credentials": "c66a355353fe41ce951ed77b7fe988bb",
  "ws_uri": "wss://nowhere"
}
```

---
## websocket

`wss://goloom.strm.yandex.net/join`

a uid is generated for each interaction (s2c | c2s). if clients sends a new uid, server responds with an ack; if server
sends a new uid, client responds with an ack. it seems like there are no exceptions to this

```json5
// send
{
  "uid": "c089e5a6-e764-4548-aceb-3aee951dfb32",
  "hello": {
    "participantMeta": {
      "name": "cool_username",
      "role": "SPEAKER",
      "description": "",
      "sendAudio": false,
      "sendVideo": false
    },
    "participantAttributes": {
      "name": "cool_username",
      "role": "SPEAKER",
      "description": ""
    },
    "sendAudio": false,
    "sendVideo": false,
    "sendSharing": false,
    "participantId": "f72c8dc2-6ac6-43de-819a-06e3c3f0227c",
    "roomId": "d48791a2-07d5-4d3b-88d9-d82557bd56ac",
    "serviceName": "telemost",
    "credentials": "c66a355353fe41ce951ed77b7fe988bb",
    "capabilitiesOffer": {
      "offerAnswerMode": [
        "SEPARATE"
      ],
      "initialSubscriberOffer": [
        "ON_HELLO"
      ],
      "slotsMode": [
        "FROM_CONTROLLER"
      ],
      "simulcastMode": [
        "DISABLED",
        "STATIC"
      ],
      "selfVadStatus": [
        "FROM_SERVER",
        "FROM_CLIENT"
      ],
      "dataChannelSharing": [
        "TO_RTP"
      ],
      "videoEncoderConfig": [
        "NO_CONFIG",
        "ONLY_INIT_CONFIG",
        "RUNTIME_CONFIG"
      ],
      "dataChannelVideoCodec": [
        "VP8",
        "UNIQUE_CODEC_FROM_TRACK_DESCRIPTION"
      ],
      "bandwidthLimitationReason": [
        "BANDWIDTH_REASON_DISABLED",
        "BANDWIDTH_REASON_ENABLED"
      ],
      "sdkDefaultDeviceManagement": [
        "SDK_DEFAULT_DEVICE_MANAGEMENT_DISABLED",
        "SDK_DEFAULT_DEVICE_MANAGEMENT_ENABLED"
      ],
      "joinOrderLayout": [
        "JOIN_ORDER_LAYOUT_DISABLED",
        "JOIN_ORDER_LAYOUT_ENABLED"
      ],
      "pinLayout": [
        "PIN_LAYOUT_DISABLED"
      ],
      "sendSelfViewVideoSlot": [
        "SEND_SELF_VIEW_VIDEO_SLOT_DISABLED",
        "SEND_SELF_VIEW_VIDEO_SLOT_ENABLED"
      ],
      "serverLayoutTransition": [
        "SERVER_LAYOUT_TRANSITION_DISABLED"
      ],
      "sdkPublisherOptimizeBitrate": [
        "SDK_PUBLISHER_OPTIMIZE_BITRATE_DISABLED",
        "SDK_PUBLISHER_OPTIMIZE_BITRATE_FULL",
        "SDK_PUBLISHER_OPTIMIZE_BITRATE_ONLY_SELF"
      ],
      "sdkNetworkLostDetection": [
        "SDK_NETWORK_LOST_DETECTION_DISABLED"
      ],
      "sdkNetworkPathMonitor": [
        "SDK_NETWORK_PATH_MONITOR_DISABLED"
      ],
      "publisherVp9": [
        "PUBLISH_VP9_DISABLED",
        "PUBLISH_VP9_ENABLED"
      ],
      "svcMode": [
        "SVC_MODE_DISABLED",
        "SVC_MODE_L3T3",
        "SVC_MODE_L3T3_KEY"
      ],
      "subscriberOfferAsyncAck": [
        "SUBSCRIBER_OFFER_ASYNC_ACK_DISABLED",
        "SUBSCRIBER_OFFER_ASYNC_ACK_ENABLED"
      ],
      "androidBluetoothRoutingFix": [
        "ANDROID_BLUETOOTH_ROUTING_FIX_DISABLED"
      ],
      "fixedIceCandidatesPoolSize": [
        "FIXED_ICE_CANDIDATES_POOL_SIZE_DISABLED"
      ],
      "sdkAndroidTelecomIntegration": [
        "SDK_ANDROID_TELECOM_INTEGRATION_DISABLED"
      ],
      "setActiveCodecsMode": [
        "SET_ACTIVE_CODECS_MODE_DISABLED",
        "SET_ACTIVE_CODECS_MODE_VIDEO_ONLY"
      ],
      "subscriberDtlsPassiveMode": [
        "SUBSCRIBER_DTLS_PASSIVE_MODE_DISABLED"
      ],
      "svcModes": [
        "FALSE"
      ],
      "reportTelemetryModes": [
        "TRUE"
      ],
      "keepDefaultDevicesModes": [
        "FALSE"
      ]
    },
    "sdkInfo": {
      "implementation": "browser",
      "version": "5.22.0",
      "userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:147.0) Gecko/20100101 Firefox/147.0",
      "hwConcurrency": 8
    },
    "sdkInitializationId": "8d52ced4-bfcf-4360-9a80-80c7fb785a33",
    "disablePublisher": false,
    "disableSubscriber": false,
    "disableSubscriberAudio": false
  }
}
```

```json5
// receive
{
  "uid": "1a200042-c3a7-4a9e-a619-561ac14acf50",
  "serverHello": {
    "capabilitiesAnswer": {
      "offerAnswerMode": "SEPARATE",
      "initialSubscriberOffer": "ON_HELLO",
      "slotsMode": "FROM_CONTROLLER",
      "simulcastMode": "DISABLED",
      "selfVadStatus": "FROM_SERVER",
      "dataChannelSharing": "TO_RTP",
      "videoEncoderConfig": "NO_CONFIG",
      "dataChannelVideoCodec": "UNIQUE_CODEC_FROM_TRACK_DESCRIPTION",
      "bandwidthLimitationReason": "BANDWIDTH_REASON_ENABLED",
      "serverLayoutTransition": "SERVER_LAYOUT_TRANSITION_DISABLED",
      "pinLayout": "PIN_LAYOUT_DISABLED",
      "joinOrderLayout": "JOIN_ORDER_LAYOUT_ENABLED",
      "sendSelfViewVideoSlot": "SEND_SELF_VIEW_VIDEO_SLOT_ENABLED",
      "sdkDefaultDeviceManagement": "SDK_DEFAULT_DEVICE_MANAGEMENT_ENABLED",
      "sdkPublisherOptimizeBitrate": "SDK_PUBLISHER_OPTIMIZE_BITRATE_FULL",
      "sdkNetworkPathMonitor": "SDK_NETWORK_PATH_MONITOR_DISABLED",
      "publisherVp9": "PUBLISH_VP9_ENABLED",
      "svcMode": "SVC_MODE_L3T3_KEY",
      "sdkNetworkLostDetection": "SDK_NETWORK_LOST_DETECTION_DISABLED",
      "fixedIceCandidatesPoolSize": "FIXED_ICE_CANDIDATES_POOL_SIZE_DISABLED",
      "subscriberOfferAsyncAck": "SUBSCRIBER_OFFER_ASYNC_ACK_DISABLED",
      "androidBluetoothRoutingFix": "ANDROID_BLUETOOTH_ROUTING_FIX_DISABLED",
      "sdkAndroidTelecomIntegration": "SDK_ANDROID_TELECOM_INTEGRATION_DISABLED",
      "setActiveCodecsMode": "SET_ACTIVE_CODECS_MODE_DISABLED",
      "subscriberDtlsPassiveMode": "SUBSCRIBER_DTLS_PASSIVE_MODE_DISABLED",
      "publisherOpusLowBitrate": "PUBLISHER_OPUS_LOW_BITRATE_DISABLED",
      "publisherOpusDred": "PUBLISHER_OPUS_DRED_DISABLED",
      "sdkAndroidDestroySessionOnTaskRemoved": "SDK_ANDROID_DESTROY_SESSION_ON_TASK_REMOVED_ENABLED"
    },
    "servingComponents": [
      {
        "type": "BORDER",
        "host": "strm-border-production-9.sas.yp-c.yandex.net",
        "version": "r18743381"
      },
      {
        "type": "WEBRTC_SERVER",
        "host": "strm-sfu-production-8b-16.sas.yp-c.yandex.net",
        "version": "r18728127"
      },
      {
        "type": "CONTROLLER",
        "host": "strm-roomcontroller-production-8-7.klg.yp-c.yandex.net",
        "version": "r18372297"
      }
    ],
    "sessionSecret": "e24eb25c-a4b1-45d9-8412-fe85c18035e9",
    "vadConfig": {
      "probabilityThreshold": 0.8,
      "debounceTimeMs": 5000,
      "activateSampleSize": 5,
      "deactivateSampleSize": 10
    },
    "sfuPeerInitializationId": "fe3b28a8-28a2-4b33-a145-563d6b990c65",
    "rtcConfiguration": {
      "iceServers": [
        {
          "urls": [
            "stun:turn.tel.yandex.net",
            "stun:stun.rtc.yandex.net"
          ],
          "credential": "",
          "username": ""
        },
        {
          "urls": [
            "turn:turn.tel.yandex.net:443"
          ],
          "credential": "K5O30K2H/I36VYpTZ7aloxs3V7o=",
          "username": "1771513226:SJoV:96339cfc-3e9f-44a2-b2f4-1a9fd8038401"
        },
        {
          "urls": [
            "turn:turn.tel.yandex.net:443"
          ],
          "credential": "MQINenT2ddddLnh51KwA5gVidIk=",
          "username": "1771513226:UJRy:96339cfc-3e9f-44a2-b2f4-1a9fd8038401"
        },
        {
          "urls": [
            "turn:turn.tel.yandex.net:443?transport=tcp"
          ],
          "credential": "fk5tOUcwx2lFvubOLIbkbNHpngM=",
          "username": "1771513226:8i5v:96339cfc-3e9f-44a2-b2f4-1a9fd8038401"
        }
      ]
    },
    "logEndpoint": "",
    "videoEncoderConfig": null,
    "sdkFeatureFlags": {
      "enableOpusDtx": false
    },
    "soundProcessingConfiguration": {
      "dfConfiguration": {
        "maxSnrForErb": 30,
        "maxSnrForDf": 30,
        "maxSnrForZeroOutput": -10,
        "minChunkPower": 1e-7,
        "minHighFreqChunkRms": 0,
        "modelVersion": ""
      }
    },
    "pingPongConfiguration": {
      "pingInterval": 5000,
      "ackTimeout": 9000
    },
    "telemetryConfiguration": {
      "sendingInterval": 20000
    },
    "videoLayersConfiguration": {
      "l1": {
        "low": {
          "bitrate": 1000000
        },
        "startBitrate": 1000000
      },
      "l2": {
        "low": {
          "bitrate": 120000
        },
        "med": {
          "bitrate": 360000
        },
        "startBitrate": 0
      },
      "l3": {
        "low": {
          "bitrate": 120000
        },
        "med": {
          "bitrate": 360000
        },
        "hi": {
          "bitrate": 800000
        },
        "startBitrate": 0
      },
      "l4": {
        "low": {
          "bitrate": 120000
        },
        "med": {
          "bitrate": 360000
        },
        "hi": {
          "bitrate": 800000
        },
        "ultra": {
          "bitrate": 1000000
        },
        "startBitrate": 0
      }
    },
    "fourKSharingConfiguration": {
      "defaultContentHint": "detail",
      "minBitrate": 300000,
      "maxBitrate": 2000000,
      "minFramerate": 8,
      "maxFramerate": 30,
      "bufferedAmountLowThreshold": 0
    },
    "excludeFromExperiments": false
  }
}
```

a bunch of webRtcIceCandidates:

```json5
// send
{
  "uid": "8f0a08ef-31d2-422e-b029-3ab9b1d1ded3",
  "webrtcIceCandidate": {
    "candidate": "candidate:0 1 UDP 2122252543 172.25.48.1 55689 typ host",
    "sdpMid": "0",
    "usernameFragment": "fb7af1ba",
    "sdpMlineIndex": 0,
    "target": "PUBLISHER",
    "pcSeq": 1
  }
}
```

```json5
// receive
{
  "uid": "e5538ad7-a082-4885-bc08-4f1ce8ef664e",
  "subscriberSdpOffer": {
    "pcSeq": 1,
    "sdp": "..."
  }
}
```

```json5
// send
{
  "uid": "e69c4fc0-4164-4265-a58f-83557f7f6849",
  "subscriberSdpAnswer": {
    "sdp": "...",
    "pcSeq": 1
  }
}
```

also some webRtcIceCandidates. maybe were supposed to be with the previous ones, but the server offer split the messages

```json5
// send
{
  "uid": "50d2967a-cff3-4714-a029-51a4e5ed58a6",
  "publisherSdpOffer": {
    "pcSeq": 1,
    "sdp": "...",
    "tracks": [
      {
        "mid": "0",
        "transceiverMid": "0",
        "kind": "AUDIO",
        "priority": 0,
        "label": "Microphone Array (Realtek(R) Audio)",
        "codecs": {},
        "groupId": 1,
        "description": ""
      }
    ]
  }
}
```

```json5
// receive
{
  "uid": "e4e9df30-6077-4746-82f6-d604494fce92",
  "publisherSdpAnswer": {
    "pcSeq": 1,
    "sdp": "..."
  }
}
```

two webrtcIceCandidates from the server:

```json5
// receive
{
  "uid": "354b957b-ec1a-41f3-baef-40f964701005",
  "webrtcIceCandidate": {
    "pcSeq": 1,
    "target": "PUBLISHER",
    "candidate": "candidate:2129394685 1 udp 2130706431 2a02:6b8:c23:4188:0:5966:260c:0 22853 typ host ufrag fZBqsNUjhynjlcgo",
    "sdpMid": "0",
    "sdpMlineIndex": 0
  }
}
```

there are more webrtcIceCandidate exchanges over the next couple of seconds or so

```json5
// send
{
  "uid": "4c961036-2b9c-4b7a-a4bd-4b61160d3561",
  "sdkCodecsInfo": {
    "vp8": {
      "supported": "CODEC_FEATURE_SUPPORTED",
      "hwDecode": "CODEC_FEATURE_SUPPORTED",
      "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
      "isoString": "vp8"
    },
    "vp9": [
      {
        "supported": "CODEC_FEATURE_SUPPORTED",
        "hwDecode": "CODEC_FEATURE_SUPPORTED",
        "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
        "isoString": "vp09.00.51.08"
      },
      {
        "supported": "CODEC_FEATURE_SUPPORTED",
        "hwDecode": "CODEC_FEATURE_SUPPORTED",
        "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
        "isoString": "vp09.02.51.10"
      },
      {
        "supported": "CODEC_FEATURE_SUPPORTED",
        "hwDecode": "CODEC_FEATURE_SUPPORTED",
        "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
        "isoString": "vp09.02.51.12"
      }
    ],
    "av1": [
      {
        "supported": "CODEC_FEATURE_SUPPORTED",
        "hwDecode": "CODEC_FEATURE_SUPPORTED",
        "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
        "isoString": "av01.0.04M.08"
      },
      {
        "supported": "CODEC_FEATURE_SUPPORTED",
        "hwDecode": "CODEC_FEATURE_SUPPORTED",
        "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
        "isoString": "av01.0.04M.10"
      },
      {
        "supported": "CODEC_FEATURE_SUPPORTED",
        "hwDecode": "CODEC_FEATURE_SUPPORTED",
        "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
        "isoString": "av01.0.05M.08"
      },
      {
        "supported": "CODEC_FEATURE_SUPPORTED",
        "hwDecode": "CODEC_FEATURE_SUPPORTED",
        "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
        "isoString": "av01.0.05M.10"
      }
    ],
    "h264": [
      {
        "supported": "CODEC_FEATURE_SUPPORTED",
        "hwDecode": "CODEC_FEATURE_SUPPORTED",
        "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
        "isoString": "avc1.42e01f"
      },
      {
        "supported": "CODEC_FEATURE_SUPPORTED",
        "hwDecode": "CODEC_FEATURE_SUPPORTED",
        "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
        "isoString": "avc1.42001f"
      },
      {
        "supported": "CODEC_FEATURE_SUPPORTED",
        "hwDecode": "CODEC_FEATURE_SUPPORTED",
        "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
        "isoString": "avc1.4d001f"
      },
      {
        "supported": "CODEC_FEATURE_SUPPORTED",
        "hwDecode": "CODEC_FEATURE_SUPPORTED",
        "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
        "isoString": "avc1.640034"
      },
      {
        "supported": "CODEC_FEATURE_SUPPORTED",
        "hwDecode": "CODEC_FEATURE_SUPPORTED",
        "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
        "isoString": "avc1.420034"
      },
      {
        "supported": "CODEC_FEATURE_SUPPORTED",
        "hwDecode": "CODEC_FEATURE_SUPPORTED",
        "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
        "isoString": "avc1.42e034"
      },
      {
        "supported": "CODEC_FEATURE_SUPPORTED",
        "hwDecode": "CODEC_FEATURE_SUPPORTED",
        "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
        "isoString": "avc1.4d0034"
      },
      {
        "supported": "CODEC_FEATURE_SUPPORTED",
        "hwDecode": "CODEC_FEATURE_SUPPORTED",
        "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
        "isoString": "avc1.640c1f"
      },
      {
        "supported": "CODEC_FEATURE_SUPPORTED",
        "hwDecode": "CODEC_FEATURE_SUPPORTED",
        "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
        "isoString": "avc1.640020"
      },
      {
        "supported": "CODEC_FEATURE_SUPPORTED",
        "hwDecode": "CODEC_FEATURE_SUPPORTED",
        "hwEncode": "CODEC_FEATURE_NOT_SUPPORTED",
        "isoString": "avc1.64001f"
      }
    ]
  }
}
```

some slot configuration. these seem to fire when you change screen dimensions. client sends setSlots, server sends
slotsConfigs:

```json5
// send
{
  "uid": "2244ab2a-ffed-4926-b710-b4b6e13d4954",
  "setSlots": {
    "slots": [
      {
        "width": 368,
        "height": 207
      },
      // ...
      {
        "width": 176,
        "height": 99
      }
    ],
    "audioSlotsCount": 0,
    "key": 2,
    "shutdownAllVideo": null,
    "withSelfView": true,
    "selfViewVisibility": "ON_LOADING_THEN_SHOW",
    "gridConfig": {}
  }
}
```

```json5
// receive
{
  "uid": "a5eeb35d-61f9-4d51-ac5a-927b4e10c1e4",
  "slotsConfig": {
    "key": 2,
    "videoSlots": [],
    "audioSlots": [],
    "prevSlots": [],
    "slots": [
      {
        "selfView": {},
        "vad": false,
        "pinned": false,
        "label": ""
      },
      // ...
      {
        "empty": {},
        "vad": false,
        "pinned": false,
        "label": ""
      }
    ],
    "nextSlots": [],
    "offset": 0,
    "gridConfig": {}
  }
}
```

```json5
// receive
{
  "uid": "14bf0db2-d0d0-4f6a-93d8-f82db0244f9d",
  "selfQualityReport": {
    "networkScore": "EXCELLENT"
  }
}
```

somewhere there was an updateDescription too, but it was empty so i left it out
only pings and telemetry are left now. telemetry is so big i don't want to paste it here

---
## get state

every 30 seconds

```http request
POST
    https://cloud-api.yandex.ru
    /telemost_front/v2/telemost/conferences
    /https%3A%2F%2Ftelemost.yandex.ru%2Fj%2F69080743110758%3F
    /request-states
```

first request:
```json5
// send
{
  "peers": [
    {
      "peer_id": "f72c8dc2-6ac6-43de-819a-06e3c3f0227c"
    }
  ],
  "permissions": {},
  "conference": {
    "version": -1
  }
}
```

```json5
// receive
{
  "permissions": {
    "version": 0,
    "public_role_permissions": [
      {
        "role": "OWNER",
        "allowed": [
          "cloud_recording",
          "camera",
          "microphone",
          "close_room",
          "summarization",
          "summarization_receive",
          "desktop",
          "calendar_summarization_receive",
          "recording"
        ]
      },
      {
        "role": "ADMIN",
        "allowed": [
          "cloud_recording",
          "camera",
          "microphone",
          "close_room",
          "summarization",
          "summarization_receive",
          "desktop",
          "recording"
        ]
      },
      {
        "role": "MEMBER",
        "allowed": [
          "camera",
          "microphone",
          "cloud_recording",
          "desktop",
          "recording"
        ]
      }
    ],
    "personal_allowed": [
      "reaction",
      "camera_enable_default",
      "microphone_enable_default"
    ]
  },
  "peers": [
    {
      "peer_id": "c7f9a1a6-5522-4950-b5c3-1197b2bf3c2c",
      "peer_type": "USER",
      "version": 0,
      "state": {
        "user_data": {
          "role": "MEMBER",
          "uid": "1678684838",
          "mssngr_guid": "6c1a6104-51a8-9e92-56ba-d6928f847c97",
          "display_name": "Saveliy Sidorov",
          "avatar_url": "https://avatars.mds.yandex.net/get-yapic/0/0-0/islands-300",
          "is_default_avatar": true,
          "avatar_placeholder": {
            "background_color": "#b948f6",
            "text_color": "#fff",
            "abbreviation": "SS"
          }
        }
      },
      "public_permissions_override": null
    }
  ],
  "conference": {
    "version": 1,
    "state": {
      "local_recording_allowed": true,
      "cloud_recording_allowed": false,
      "chat_allowed": true,
      "control_allowed": true,
      "broadcast_allowed": false,
      "broadcast_feature_enabled": false,
      "access_restriction_organization_allowed": false,
      "chat_path": "0/22/fd6676eb-4196-4296-a1ce-635c1bcbe366",
      "access_level": "PUBLIC",
      "summarization_status": "TURNED_OFF",
      "cloud_recording_status": "TURNED_OFF"
    }
  }
}
```

after that all requests:

```json5
// send
{
  "peers": [
    {
      "peer_id": "c7f9a1a6-5522-4950-b5c3-1197b2bf3c2c",
      "version": 0
    }
  ],
  "conference": {
    "version": 1
  },
  "permissions": {
    "version": 0
  }
}
```

```json5
// receive
{
  "peers": []
}
```

---

# when a peer joins the current call (requests continue from the last section)
## ws:

```json5
// receive
{
  "uid": "8e0ad11a-ebb9-412b-b618-8d534ffd2991",
  "upsertDescription": {
    "description": [
      {
        "id": "b06b0c87-e031-4983-a503-a63af1af4285",
        "meta": {
          "name": "cool_peer2",
          "role": "SPEAKER",
          "description": "",
          "sendAudio": false,
          "sendVideo": false
        },
        "participantAttributes": {},
        "sendAudio": false,
        "sendVideo": false,
        "sendSharing": false,
        "hideFromParticipantsList": false,
        "networkScore": "NETWORK_QUALITY_SCORE_UNSPECIFIED",
        "connectionType": "CONNECTION_TYPE_SDK"
      }
    ]
  }
}
```

```json5
// receive
{
  "uid": "ca24c540-d702-4310-a413-a291cf669fde",
  "upsertParticipantsQualityReport": {
    "participantsQualityReport": [
      {
        "participantId": "b06b0c87-e031-4983-a503-a63af1af4285",
        "networkScore": "EXCELLENT"
      }
    ]
  }
}
```

---
## states

request-states still returns []

```http request
POST
    https://cloud-api.yandex.ru
    /telemost_front/v2/telemost/conferences
    /https%3A%2F%2Ftelemost.yandex.ru%2Fj%2F69080743110758%3F
    /request-states
```

```json5
{
  "peers": []
}
```

---

# from cool_peer2's perspective (joining a call where there's a peer already)

```http request
GET
    https://cloud-api.yandex.ru
    /telemost_front/v2/telemost/conferences
    /https%3A%2F%2Ftelemost.yandex.ru%2Fj%2F69080743110758
    /connection?next_gen_media_platform_allowed=true
    &display_name=cool_peer2&waiting_room_supported=true
```

```json5
{
  "connection_type": "CONFERENCE",
  "uri": "https://telemost.yandex.ru/j/69080743110758",
  "room_id": "d48791a2-07d5-4d3b-88d9-d82557bd56ac",
  "safe_room_id": "d48791a2-07d5-4d3b-88d9-d82557bd56ac",
  "peer_id": "b06b0c87-e031-4983-a503-a63af1af4285",
  "session_id": "4b2c4d25-5fb8-4710-9912-2ddd05332c15",
  "peer_session_id": "234a81cf-4cb8-4266-a550-4e2864051d11",
  "credentials": "86314c7d3fd945cfaea7e529d31117c6",
  "client_configuration": {
    "cloud_recording_available": false,
    "wait_time_to_reconnect_ms": 4820,
    "media_server_url": "wss://goloom.strm.yandex.net/join",
    "service_name": "telemost",
    "alice_pro_enabled": true,
    "summarization_available": false,
    "new_grid_mobile": false,
    "extended_role_permissions_enabled": false,
    "calendar_summarization_receive_available": false,
    "reactions_available": true,
    "join_url_hidden": false,
    "state_check_interval_seconds": 30,
    "goloom_session_open_ms": 120000,
    "new_grid": true,
    "waiting_room_peers_refresh_interval_ms": 10000,
    "report_problem_button_available": false,
    "ice_servers": [
      {
        "urls": [
          "stun:stun.rtc.yandex.net:3478"
        ]
      }
    ]
  },
  "conference_state": {
    "local_recording_allowed": true,
    "cloud_recording_allowed": false,
    "chat_allowed": true,
    "control_allowed": true,
    "broadcast_allowed": false,
    "broadcast_feature_enabled": false,
    "access_restriction_organization_allowed": false,
    "chat_path": "0/22/577ba3bb-9b23-4bd4-b148-db18916582eb",
    "access_level": "PUBLIC",
    "summarization_status": "TURNED_OFF",
    "cloud_recording_status": "TURNED_OFF"
  },
  "media_platform": "GOLOOM",
  "is_legal_entity": false,
  "waiting_room_available": false,
  "expiration_time": 1771509742987,
  "conference_limit": 40,
  "ws_uri": "wss://nowhere"
}
```

---

## ws:

```json5
// send
{
  "uid": "4f0612d7-cda6-4322-909c-ade9f1f133fc",
  "hello": {
    "participantMeta": {
      "name": "cool_peer2",
      "role": "SPEAKER",
      "description": "",
      "sendAudio": false,
      "sendVideo": false
    },
    "participantAttributes": {
      "name": "cool_peer2",
      "role": "SPEAKER",
      "description": ""
    },
    "sendAudio": false,
    "sendVideo": false,
    "sendSharing": false,
    "participantId": "b06b0c87-e031-4983-a503-a63af1af4285",
    "roomId": "d48791a2-07d5-4d3b-88d9-d82557bd56ac",
    "serviceName": "telemost",
    "credentials": "86314c7d3fd945cfaea7e529d31117c6",
    "capabilitiesOffer": {
      "offerAnswerMode": [
        "SEPARATE"
      ],
      "initialSubscriberOffer": [
        "ON_HELLO"
      ],
      "slotsMode": [
        "FROM_CONTROLLER"
      ],
      "simulcastMode": [
        "DISABLED",
        "STATIC"
      ],
      "selfVadStatus": [
        "FROM_SERVER",
        "FROM_CLIENT"
      ],
      "dataChannelSharing": [
        "TO_RTP"
      ],
      "videoEncoderConfig": [
        "NO_CONFIG",
        "ONLY_INIT_CONFIG",
        "RUNTIME_CONFIG"
      ],
      "dataChannelVideoCodec": [
        "VP8",
        "UNIQUE_CODEC_FROM_TRACK_DESCRIPTION"
      ],
      "bandwidthLimitationReason": [
        "BANDWIDTH_REASON_DISABLED",
        "BANDWIDTH_REASON_ENABLED"
      ],
      "sdkDefaultDeviceManagement": [
        "SDK_DEFAULT_DEVICE_MANAGEMENT_DISABLED",
        "SDK_DEFAULT_DEVICE_MANAGEMENT_ENABLED"
      ],
      "joinOrderLayout": [
        "JOIN_ORDER_LAYOUT_DISABLED",
        "JOIN_ORDER_LAYOUT_ENABLED"
      ],
      "pinLayout": [
        "PIN_LAYOUT_DISABLED"
      ],
      "sendSelfViewVideoSlot": [
        "SEND_SELF_VIEW_VIDEO_SLOT_DISABLED",
        "SEND_SELF_VIEW_VIDEO_SLOT_ENABLED"
      ],
      "serverLayoutTransition": [
        "SERVER_LAYOUT_TRANSITION_DISABLED"
      ],
      "sdkPublisherOptimizeBitrate": [
        "SDK_PUBLISHER_OPTIMIZE_BITRATE_DISABLED",
        "SDK_PUBLISHER_OPTIMIZE_BITRATE_FULL",
        "SDK_PUBLISHER_OPTIMIZE_BITRATE_ONLY_SELF"
      ],
      "sdkNetworkLostDetection": [
        "SDK_NETWORK_LOST_DETECTION_DISABLED"
      ],
      "sdkNetworkPathMonitor": [
        "SDK_NETWORK_PATH_MONITOR_DISABLED"
      ],
      "publisherVp9": [
        "PUBLISH_VP9_DISABLED",
        "PUBLISH_VP9_ENABLED"
      ],
      "svcMode": [
        "SVC_MODE_DISABLED",
        "SVC_MODE_L3T3",
        "SVC_MODE_L3T3_KEY"
      ],
      "subscriberOfferAsyncAck": [
        "SUBSCRIBER_OFFER_ASYNC_ACK_DISABLED",
        "SUBSCRIBER_OFFER_ASYNC_ACK_ENABLED"
      ],
      "androidBluetoothRoutingFix": [
        "ANDROID_BLUETOOTH_ROUTING_FIX_DISABLED"
      ],
      "fixedIceCandidatesPoolSize": [
        "FIXED_ICE_CANDIDATES_POOL_SIZE_DISABLED"
      ],
      "sdkAndroidTelecomIntegration": [
        "SDK_ANDROID_TELECOM_INTEGRATION_DISABLED"
      ],
      "setActiveCodecsMode": [
        "SET_ACTIVE_CODECS_MODE_DISABLED",
        "SET_ACTIVE_CODECS_MODE_VIDEO_ONLY"
      ],
      "subscriberDtlsPassiveMode": [
        "SUBSCRIBER_DTLS_PASSIVE_MODE_DISABLED"
      ],
      "svcModes": [
        "FALSE"
      ],
      "reportTelemetryModes": [
        "TRUE"
      ],
      "keepDefaultDevicesModes": [
        "FALSE"
      ]
    },
    "sdkInfo": {
      "implementation": "browser",
      "version": "5.22.0",
      "userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:147.0) Gecko/20100101 Firefox/147.0",
      "hwConcurrency": 8
    },
    "sdkInitializationId": "30c7e1b8-fc82-4f88-b4b2-b6cd3a485453",
    "disablePublisher": false,
    "disableSubscriber": false,
    "disableSubscriberAudio": false
  }
}
```

```json5
// receive
{
  "uid": "77e83b68-8659-4ba2-a720-2ba1913c8de8",
  "serverHello": {
    "capabilitiesAnswer": {
      "offerAnswerMode": "SEPARATE",
      "initialSubscriberOffer": "ON_HELLO",
      "slotsMode": "FROM_CONTROLLER",
      "simulcastMode": "DISABLED",
      "selfVadStatus": "FROM_SERVER",
      "dataChannelSharing": "TO_RTP",
      "videoEncoderConfig": "NO_CONFIG",
      "dataChannelVideoCodec": "UNIQUE_CODEC_FROM_TRACK_DESCRIPTION",
      "bandwidthLimitationReason": "BANDWIDTH_REASON_ENABLED",
      "serverLayoutTransition": "SERVER_LAYOUT_TRANSITION_DISABLED",
      "pinLayout": "PIN_LAYOUT_DISABLED",
      "joinOrderLayout": "JOIN_ORDER_LAYOUT_ENABLED",
      "sendSelfViewVideoSlot": "SEND_SELF_VIEW_VIDEO_SLOT_ENABLED",
      "sdkDefaultDeviceManagement": "SDK_DEFAULT_DEVICE_MANAGEMENT_ENABLED",
      "sdkPublisherOptimizeBitrate": "SDK_PUBLISHER_OPTIMIZE_BITRATE_FULL",
      "sdkNetworkPathMonitor": "SDK_NETWORK_PATH_MONITOR_DISABLED",
      "publisherVp9": "PUBLISH_VP9_ENABLED",
      "svcMode": "SVC_MODE_L3T3_KEY",
      "sdkNetworkLostDetection": "SDK_NETWORK_LOST_DETECTION_DISABLED",
      "fixedIceCandidatesPoolSize": "FIXED_ICE_CANDIDATES_POOL_SIZE_DISABLED",
      "subscriberOfferAsyncAck": "SUBSCRIBER_OFFER_ASYNC_ACK_DISABLED",
      "androidBluetoothRoutingFix": "ANDROID_BLUETOOTH_ROUTING_FIX_DISABLED",
      "sdkAndroidTelecomIntegration": "SDK_ANDROID_TELECOM_INTEGRATION_DISABLED",
      "setActiveCodecsMode": "SET_ACTIVE_CODECS_MODE_DISABLED",
      "subscriberDtlsPassiveMode": "SUBSCRIBER_DTLS_PASSIVE_MODE_DISABLED",
      "publisherOpusLowBitrate": "PUBLISHER_OPUS_LOW_BITRATE_DISABLED",
      "publisherOpusDred": "PUBLISHER_OPUS_DRED_DISABLED",
      "sdkAndroidDestroySessionOnTaskRemoved": "SDK_ANDROID_DESTROY_SESSION_ON_TASK_REMOVED_ENABLED"
    },
    "servingComponents": [
      {
        "type": "BORDER",
        "host": "strm-border-production-11.vla.yp-c.yandex.net",
        "version": "r18743381"
      },
      {
        "type": "WEBRTC_SERVER",
        "host": "strm-sfu-production-8b-16.sas.yp-c.yandex.net",
        "version": "r18728127"
      },
      {
        "type": "CONTROLLER",
        "host": "strm-roomcontroller-production-8-7.klg.yp-c.yandex.net",
        "version": "r18372297"
      }
    ],
    "sessionSecret": "69604f1c-9abd-47ff-9cb0-e02bf109badb",
    "vadConfig": {
      "probabilityThreshold": 0.8,
      "debounceTimeMs": 5000,
      "activateSampleSize": 5,
      "deactivateSampleSize": 10
    },
    "sfuPeerInitializationId": "897b4977-b8a7-4131-9a08-49575607a5fc",
    "rtcConfiguration": {
      "iceServers": [
        {
          "urls": [
            "stun:turn.tel.yandex.net",
            "stun:stun.rtc.yandex.net"
          ],
          "credential": "",
          "username": ""
        },
        {
          "urls": [
            "turn:turn.tel.yandex.net:443"
          ],
          "credential": "H7Za+nQfZLteogIZv61CD+vXlIE=",
          "username": "1771515389:PVG8:a6dd67f7-233d-4613-8c9f-ca6465d0baa2"
        },
        {
          "urls": [
            "turn:turn.tel.yandex.net:443"
          ],
          "credential": "YZAR0ttDEpESRDEQDXfEFbV8mg4=",
          "username": "1771515389:C2BK:a6dd67f7-233d-4613-8c9f-ca6465d0baa2"
        },
        {
          "urls": [
            "turn:turn.tel.yandex.net:443?transport=tcp"
          ],
          "credential": "dG5SdHkJ5HOIh/zCXkDLsBDrBTs=",
          "username": "1771515389:Jkgs:a6dd67f7-233d-4613-8c9f-ca6465d0baa2"
        }
      ]
    },
    "logEndpoint": "",
    "videoEncoderConfig": null,
    "sdkFeatureFlags": {
      "enableOpusDtx": false
    },
    "soundProcessingConfiguration": {
      "dfConfiguration": {
        "maxSnrForErb": 30,
        "maxSnrForDf": 30,
        "maxSnrForZeroOutput": -10,
        "minChunkPower": 1e-7,
        "minHighFreqChunkRms": 0,
        "modelVersion": ""
      }
    },
    "pingPongConfiguration": {
      "pingInterval": 5000,
      "ackTimeout": 9000
    },
    "telemetryConfiguration": {
      "sendingInterval": 20000
    },
    "videoLayersConfiguration": {
      "l1": {
        "low": {
          "bitrate": 1000000
        },
        "startBitrate": 1000000
      },
      "l2": {
        "low": {
          "bitrate": 120000
        },
        "med": {
          "bitrate": 360000
        },
        "startBitrate": 0
      },
      "l3": {
        "low": {
          "bitrate": 120000
        },
        "med": {
          "bitrate": 360000
        },
        "hi": {
          "bitrate": 800000
        },
        "startBitrate": 0
      },
      "l4": {
        "low": {
          "bitrate": 120000
        },
        "med": {
          "bitrate": 360000
        },
        "hi": {
          "bitrate": 800000
        },
        "ultra": {
          "bitrate": 1000000
        },
        "startBitrate": 0
      }
    },
    "fourKSharingConfiguration": {
      "defaultContentHint": "detail",
      "minBitrate": 300000,
      "maxBitrate": 2000000,
      "minFramerate": 8,
      "maxFramerate": 30,
      "bufferedAmountLowThreshold": 0
    },
    "excludeFromExperiments": false
  }
}
```

```json5
// send
{
  "uid": "2844e922-79ab-489f-8601-4fb4b6288d02",
  "publisherSdpOffer": {
    "pcSeq": 1,
    "sdp": "...",
    "tracks": [
      {
        "mid": "0",
        "transceiverMid": "0",
        "kind": "AUDIO",
        "priority": 0,
        "label": "Microphone Array (Realtek(R) Audio)",
        "codecs": {},
        "groupId": 1,
        "description": ""
      }
    ]
  }
}
```

bunch of webRtcIceCandidate exchanges

```json5
// receive
{
  "uid": "07ea61c9-0669-418d-ba80-7830cbeccefe",
  "updateDescription": {
    "description": [
      {
        "id": "f72c8dc2-6ac6-43de-819a-06e3c3f0227c",
        "meta": {
          "name": "cool_username",
          "role": "SPEAKER",
          "description": "",
          "sendAudio": false,
          "sendVideo": false
        },
        "participantAttributes": {},
        "sendAudio": false,
        "sendVideo": false,
        "sendSharing": false,
        "hideFromParticipantsList": false,
        "networkScore": "EXCELLENT",
        "connectionType": "CONNECTION_TYPE_SDK"
      }
    ]
  }
}
```

```json5
// receive
{
  "uid": "53a8808b-7bae-4bc5-8f39-f7994403cfdc",
  "subscriberSdpOffer": {
    "pcSeq": 1,
    "sdp": "..."
  }
}
```

```json5
// send 
{
  "uid": "984dffe8-4f78-42f4-8814-06eff898c8e3",
  "subscriberSdpAnswer": {
    "sdp": "...",
    "pcSeq": 1
  }
}
```

```json5
// receive 
{
  "uid": "7a803e27-c95e-4057-bd99-00cbf731e42c",
  "publisherSdpAnswer": {
    "pcSeq": 1,
    "sdp": "..."
  }
}
```

then slots - nothing interesting
also sdkCodecsInfo
then another selfQualityReport from server

basically only thing that changed is the non-empty updateDescription and credentials for rtc

---

same thing with request-states.

first request:

```json5
// receive
{
  "permissions": {
    "version": 0,
    "public_role_permissions": [
      {
        "role": "MEMBER",
        "allowed": [
          "recording",
          "microphone",
          "desktop",
          "cloud_recording",
          "camera"
        ]
      },
      {
        "role": "OWNER",
        "allowed": [
          "summarization_receive",
          "recording",
          "calendar_summarization_receive",
          "close_room",
          "microphone",
          "desktop",
          "cloud_recording",
          "summarization",
          "camera"
        ]
      },
      {
        "role": "ADMIN",
        "allowed": [
          "summarization_receive",
          "recording",
          "close_room",
          "microphone",
          "desktop",
          "cloud_recording",
          "summarization",
          "camera"
        ]
      }
    ],
    "personal_allowed": [
      "microphone_enable_default",
      "camera_enable_default",
      "reaction"
    ]
  },
  "peers": [
    {
      "peer_id": "f72c8dc2-6ac6-43de-819a-06e3c3f0227c",
      "peer_type": "USER",
      "version": 2,
      "state": {
        "user_data": {
          "role": "MEMBER",
          "display_name": "cool_username",
          "avatar_placeholder": {
            "background_color": "#68c99e",
            "text_color": "#000",
            "abbreviation": "C"
          }
        }
      },
      "public_permissions_override": null
    },
    {
      "peer_id": "b06b0c87-e031-4983-a503-a63af1af4285",
      "peer_type": "USER",
      "version": 0,
      "state": {
        "user_data": {
          "role": "MEMBER",
          "display_name": "cool_peer2",
          "avatar_placeholder": {
            "background_color": "#9ce5fc",
            "text_color": "#000",
            "abbreviation": "C"
          }
        }
      },
      "public_permissions_override": null
    }
  ],
  "conference": {
    "version": 1,
    "state": {
      "local_recording_allowed": true,
      "cloud_recording_allowed": false,
      "chat_allowed": true,
      "control_allowed": true,
      "broadcast_allowed": false,
      "broadcast_feature_enabled": false,
      "access_restriction_organization_allowed": false,
      "chat_path": "0/22/577ba3bb-9b23-4bd4-b148-db18916582eb",
      "access_level": "PUBLIC",
      "summarization_status": "TURNED_OFF",
      "cloud_recording_status": "TURNED_OFF"
    }
  }
}
```
after that, peers: []