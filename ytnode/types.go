package ytnode

import "encoding/json"

type PeerId string

type InitializationHTTPData struct {
	ConnectionType      string `json:"connection_type"`
	URI                 string `json:"uri"`
	RoomID              string `json:"room_id"`
	SafeRoomID          string `json:"safe_room_id"`
	PeerID              string `json:"peer_id"`
	ClientConfiguration struct {
		CloudRecordingAvailable               bool   `json:"cloud_recording_available"`
		WaitTimeToReconnectMs                 int    `json:"wait_time_to_reconnect_ms"`
		MediaServerURL                        string `json:"media_server_url"`
		ServiceName                           string `json:"service_name"`
		AliceProEnabled                       bool   `json:"alice_pro_enabled"`
		SummarizationAvailable                bool   `json:"summarization_available"`
		NewGridMobile                         bool   `json:"new_grid_mobile"`
		ExtendedRolePermissionsEnabled        bool   `json:"extended_role_permissions_enabled"`
		CalendarSummarizationReceiveAvailable bool   `json:"calendar_summarization_receive_available"`
		ReactionsAvailable                    bool   `json:"reactions_available"`
		JoinURLHidden                         bool   `json:"join_url_hidden"`
		StateCheckIntervalSeconds             int    `json:"state_check_interval_seconds"`
		GoloomSessionOpenMs                   int    `json:"goloom_session_open_ms"`
		NewGrid                               bool   `json:"new_grid"`
		WaitingRoomPeersRefreshIntervalMs     int    `json:"waiting_room_peers_refresh_interval_ms"`
		ReportProblemButtonAvailable          bool   `json:"report_problem_button_available"`
		IceServers                            []struct {
			Urls []string `json:"urls"`
		} `json:"ice_servers"`
	} `json:"client_configuration"`
	ConferenceState struct {
		LocalRecordingAllowed                bool   `json:"local_recording_allowed"`
		CloudRecordingAllowed                bool   `json:"cloud_recording_allowed"`
		ChatAllowed                          bool   `json:"chat_allowed"`
		ControlAllowed                       bool   `json:"control_allowed"`
		BroadcastAllowed                     bool   `json:"broadcast_allowed"`
		BroadcastFeatureEnabled              bool   `json:"broadcast_feature_enabled"`
		AccessRestrictionOrganizationAllowed bool   `json:"access_restriction_organization_allowed"`
		ChatPath                             string `json:"chat_path"`
		AccessLevel                          string `json:"access_level"`
		SummarizationStatus                  string `json:"summarization_status"`
		CloudRecordingStatus                 string `json:"cloud_recording_status"`
	} `json:"conference_state"`
	MediaPlatform        string `json:"media_platform"`
	IsLegalEntity        bool   `json:"is_legal_entity"`
	SessionID            string `json:"session_id"`
	WaitingRoomAvailable bool   `json:"waiting_room_available"`
	ExpirationTime       int64  `json:"expiration_time"`
	ConferenceLimit      int    `json:"conference_limit"`
	PeerSessionID        string `json:"peer_session_id"`
	Credentials          string `json:"credentials"`
	WsURI                string `json:"ws_uri"`
}

type StatesHTTPData struct {
	Permissions struct {
		Version               int `json:"version"`
		PublicRolePermissions []struct {
			Role    string   `json:"role"`
			Allowed []string `json:"allowed"`
		} `json:"public_role_permissions"`
		PersonalAllowed []string `json:"personal_allowed"`
	} `json:"permissions"`
	Peers []struct {
		PeerID   string `json:"peer_id"`
		PeerType string `json:"peer_type"`
		Version  int    `json:"version"`
		State    struct {
			UserData struct {
				Role              string `json:"role"`
				DisplayName       string `json:"display_name"`
				AvatarPlaceholder struct {
					BackgroundColor string `json:"background_color"`
					TextColor       string `json:"text_color"`
					Abbreviation    string `json:"abbreviation"`
				} `json:"avatar_placeholder"`
			} `json:"user_data"`
		} `json:"state"`
		PublicPermissionsOverride any `json:"public_permissions_override"`
	} `json:"peers"`
	Conference struct {
		Version int `json:"version"`
		State   struct {
			LocalRecordingAllowed                bool   `json:"local_recording_allowed"`
			CloudRecordingAllowed                bool   `json:"cloud_recording_allowed"`
			ChatAllowed                          bool   `json:"chat_allowed"`
			ControlAllowed                       bool   `json:"control_allowed"`
			BroadcastAllowed                     bool   `json:"broadcast_allowed"`
			BroadcastFeatureEnabled              bool   `json:"broadcast_feature_enabled"`
			AccessRestrictionOrganizationAllowed bool   `json:"access_restriction_organization_allowed"`
			ChatPath                             string `json:"chat_path"`
			AccessLevel                          string `json:"access_level"`
			SummarizationStatus                  string `json:"summarization_status"`
			CloudRecordingStatus                 string `json:"cloud_recording_status"`
		} `json:"state"`
	} `json:"conference"`
}

type PeerWithVersion struct {
	PeerID  string `json:"peer_id"`
	Version *int   `json:"version"`
}

type ConferenceState struct {
	Version int `json:"version"`
}

type PermissionsState struct {
	Version *int `json:"version"`
}

type StatesHTTPRequestData struct {
	Peers       []PeerWithVersion `json:"peers"`
	Conference  ConferenceState   `json:"conference"`
	Permissions PermissionsState  `json:"permissions"`
}

type WSMessageIncoming struct {
	Uid                             string                     `json:"uid"`
	ServerHello                     *ServerHelloPayload        `json:"serverHello,omitempty"`
	SubscriberSdpOffer              *SdpPayload                `json:"subscriberSdpOffer,omitempty"`
	PublisherSdpAnswer              *SdpPayload                `json:"publisherSdpAnswer,omitempty"`
	WebrtcIceCandidate              *WebrtcIceCandidatePayload `json:"webrtcIceCandidate,omitempty"`
	SlotsConfig                     *SlotsConfigPayload        `json:"slotsConfig,omitempty"`
	SelfQualityReport               *SelfQualityReportPayload  `json:"selfQualityReport,omitempty"`
	UpsertParticipantsQualityReport *struct{}                  `json:"upsertParticipantsQualityReport,omitempty"`
	UpdateDescription               *UpdateDescriptionPayload  `json:"updateDescription,omitempty"`
	UpsertDescription               *UpdateDescriptionPayload  `json:"upsertDescription,omitempty"`
	RemoveDescription               *RemoveDescriptionPayload  `json:"removeDescription,omitempty"`
	VadActivity                     *struct{}                  `json:"vadActivity,omitempty"`
	Ack                             *AckPayload                `json:"ack,omitempty"`
	SlotsMeta                       *struct{}                  `json:"slotsMeta,omitempty"`

	Raw *json.RawMessage
}

type WSMessageOutgoing struct {
	Uid                             string                                  `json:"uid,omitempty"`
	Hello                           *HelloPayload                           `json:"hello,omitempty"`
	WebrtcIceCandidate              *WebrtcIceCandidatePayload              `json:"webrtcIceCandidate,omitempty"`
	SubscriberSdpAnswer             *SdpPayload                             `json:"subscriberSdpAnswer,omitempty"`
	PublisherSdpOffer               *PublisherSdpOfferPayload               `json:"publisherSdpOffer,omitempty"`
	UpdatePublisherTrackDescription *UpdatePublisherTrackDescriptionPayload `json:"updatePublisherTrackDescription,omitempty"`
	SdkCodecsInfo                   *SdkCodecsInfoPayload                   `json:"sdkCodecsInfo,omitempty"`
	SetSlots                        *SetSlotsPayload                        `json:"setSlots,omitempty"`
	UpdateMe                        *struct{}                               `json:"updateMe,omitempty"`
	Ping                            *struct{}                               `json:"ping,omitempty"`
	Ack                             *AckPayload                             `json:"ack,omitempty"`
}

// incoming payloads:

type (
	ServerHelloPayload struct {
		CapabilitiesAnswer struct {
			OfferAnswerMode                       string `json:"offerAnswerMode"`
			InitialSubscriberOffer                string `json:"initialSubscriberOffer"`
			SlotsMode                             string `json:"slotsMode"`
			SimulcastMode                         string `json:"simulcastMode"`
			SelfVadStatus                         string `json:"selfVadStatus"`
			DataChannelSharing                    string `json:"dataChannelSharing"`
			VideoEncoderConfig                    string `json:"videoEncoderConfig"`
			DataChannelVideoCodec                 string `json:"dataChannelVideoCodec"`
			BandwidthLimitationReason             string `json:"bandwidthLimitationReason"`
			ServerLayoutTransition                string `json:"serverLayoutTransition"`
			PinLayout                             string `json:"pinLayout"`
			JoinOrderLayout                       string `json:"joinOrderLayout"`
			SendSelfViewVideoSlot                 string `json:"sendSelfViewVideoSlot"`
			SdkDefaultDeviceManagement            string `json:"sdkDefaultDeviceManagement"`
			SdkPublisherOptimizeBitrate           string `json:"sdkPublisherOptimizeBitrate"`
			SdkNetworkPathMonitor                 string `json:"sdkNetworkPathMonitor"`
			PublisherVp9                          string `json:"publisherVp9"`
			SvcMode                               string `json:"svcMode"`
			SdkNetworkLostDetection               string `json:"sdkNetworkLostDetection"`
			FixedIceCandidatesPoolSize            string `json:"fixedIceCandidatesPoolSize"`
			SubscriberOfferAsyncAck               string `json:"subscriberOfferAsyncAck"`
			AndroidBluetoothRoutingFix            string `json:"androidBluetoothRoutingFix"`
			SdkAndroidTelecomIntegration          string `json:"sdkAndroidTelecomIntegration"`
			SetActiveCodecsMode                   string `json:"setActiveCodecsMode"`
			SubscriberDtlsPassiveMode             string `json:"subscriberDtlsPassiveMode"`
			PublisherOpusLowBitrate               string `json:"publisherOpusLowBitrate"`
			PublisherOpusDred                     string `json:"publisherOpusDred"`
			SdkAndroidDestroySessionOnTaskRemoved string `json:"sdkAndroidDestroySessionOnTaskRemoved"`
		} `json:"capabilitiesAnswer"`
		ServingComponents []ServingComponent `json:"servingComponents"`
		SessionSecret     string             `json:"sessionSecret"`
		VadConfig         struct {
			ProbabilityThreshold float64 `json:"probabilityThreshold"`
			DebounceTimeMs       int     `json:"debounceTimeMs"`
			ActivateSampleSize   int     `json:"activateSampleSize"`
			DeactivateSampleSize int     `json:"deactivateSampleSize"`
		} `json:"vadConfig"`
		SfuPeerInitializationID string `json:"sfuPeerInitializationId"`
		RtcConfiguration        struct {
			IceServers []IceServer `json:"iceServers"`
		} `json:"rtcConfiguration"`
		LogEndpoint        string `json:"logEndpoint"`
		VideoEncoderConfig any    `json:"videoEncoderConfig"`
		SdkFeatureFlags    struct {
			EnableOpusDtx bool `json:"enableOpusDtx"`
		} `json:"sdkFeatureFlags"`
		SoundProcessingConfiguration struct {
			DfConfiguration struct {
				MaxSnrForErb        int     `json:"maxSnrForErb"`
				MaxSnrForDf         int     `json:"maxSnrForDf"`
				MaxSnrForZeroOutput int     `json:"maxSnrForZeroOutput"`
				MinChunkPower       float64 `json:"minChunkPower"`
				MinHighFreqChunkRms int     `json:"minHighFreqChunkRms"`
				ModelVersion        string  `json:"modelVersion"`
			} `json:"dfConfiguration"`
		} `json:"soundProcessingConfiguration"`
		PingPongConfiguration struct {
			PingInterval int `json:"pingInterval"`
			AckTimeout   int `json:"ackTimeout"`
		} `json:"pingPongConfiguration"`
		TelemetryConfiguration struct {
			SendingInterval int `json:"sendingInterval"`
		} `json:"telemetryConfiguration"`
		VideoLayersConfiguration struct {
			L1 VideoLayerBitrate `json:"l1"`
			L2 VideoLayerBitrate `json:"l2"`
			L3 VideoLayerBitrate `json:"l3"`
			L4 VideoLayerBitrate `json:"l4"`
		} `json:"videoLayersConfiguration"`
		FourKSharingConfiguration struct {
			DefaultContentHint         string `json:"defaultContentHint"`
			MinBitrate                 int    `json:"minBitrate"`
			MaxBitrate                 int    `json:"maxBitrate"`
			MinFramerate               int    `json:"minFramerate"`
			MaxFramerate               int    `json:"maxFramerate"`
			BufferedAmountLowThreshold int    `json:"bufferedAmountLowThreshold"`
		} `json:"fourKSharingConfiguration"`
		ExcludeFromExperiments bool `json:"excludeFromExperiments"`
	}
	VideoLayerBitrate struct {
		Low          BitrateConfig `json:"low"`
		Med          BitrateConfig `json:"med,omitempty"`
		Hi           BitrateConfig `json:"hi,omitempty"`
		Ultra        BitrateConfig `json:"ultra,omitempty"`
		StartBitrate int           `json:"startBitrate"`
	}
	BitrateConfig struct {
		Bitrate int `json:"bitrate"`
	}
	ServingComponent struct {
		Type    string `json:"type"`
		Host    string `json:"host"`
		Version string `json:"version"`
	}
	IceServer struct {
		Urls       []string `json:"urls"`
		Credential string   `json:"credential"`
		Username   string   `json:"username"`
	}
)

type SelfQualityReportPayload struct {
	NetworkScore string `json:"networkScore"`
}

type (
	UpdateDescriptionPayload struct {
		Description []ParticipantDescription `json:"description"`
	}
	ParticipantDescription struct {
		Id                       PeerId                `json:"id"`
		Meta                     ParticipantMeta       `json:"meta"`
		ParticipantAttributes    ParticipantAttributes `json:"participantAttributes"`
		SendAudio                bool                  `json:"sendAudio"`
		SendVideo                bool                  `json:"sendVideo"`
		SendSharing              bool                  `json:"sendSharing"`
		HideFromParticipantsList bool                  `json:"hideFromParticipantsList"`
		NetworkScore             string                `json:"networkScore"`
		ConnectionType           string                `json:"connectionType"`
	}
)

type (
	SlotsConfigPayload struct {
		Key        int           `json:"key"`
		VideoSlots []interface{} `json:"videoSlots"`
		AudioSlots []interface{} `json:"audioSlots"`
		PrevSlots  []interface{} `json:"prevSlots"`
		Slots      []Slot        `json:"slots"`
		NextSlots  []interface{} `json:"nextSlots"`
		Offset     int           `json:"offset"`
		GridConfig struct{}      `json:"gridConfig"`
	}
	Slot struct {
		SelfView              *struct{}              `json:"selfView,omitempty"`
		Empty                 *struct{}              `json:"empty,omitempty"`
		Vad                   bool                   `json:"vad"`
		Pinned                bool                   `json:"pinned"`
		Label                 string                 `json:"label"`
		ParticipantVideoByMid *ParticipantVideoByMid `json:"participantVideoByMid,omitempty"`
		Participant           *Participant           `json:"participant,omitempty"`
	}
	Participant struct {
		ParticipantId PeerId `json:"participantId"`
	}
	ParticipantVideoByMid struct {
		ParticipantId    PeerId `json:"participantId"`
		Mid              string `json:"mid"`
		LimitationReason string `json:"limitationReason"`
	}
)

type RemoveDescriptionPayload struct {
	DescriptionId []PeerId `json:"descriptionId"`
}

// outgoing payloads:

type (
	HelloPayload struct {
		ParticipantMeta       ParticipantMeta       `json:"participantMeta"`
		ParticipantAttributes ParticipantAttributes `json:"participantAttributes"`
		SendAudio             bool                  `json:"sendAudio"`
		SendVideo             bool                  `json:"sendVideo"`
		SendSharing           bool                  `json:"sendSharing"`
		ParticipantID         string                `json:"participantId"`
		RoomID                string                `json:"roomId"`
		ServiceName           string                `json:"serviceName"`
		Credentials           string                `json:"credentials"`
		CapabilitiesOffer     struct {
			OfferAnswerMode              []string `json:"offerAnswerMode"`
			InitialSubscriberOffer       []string `json:"initialSubscriberOffer"`
			SlotsMode                    []string `json:"slotsMode"`
			SimulcastMode                []string `json:"simulcastMode"`
			SelfVadStatus                []string `json:"selfVadStatus"`
			DataChannelSharing           []string `json:"dataChannelSharing"`
			VideoEncoderConfig           []string `json:"videoEncoderConfig"`
			DataChannelVideoCodec        []string `json:"dataChannelVideoCodec"`
			BandwidthLimitationReason    []string `json:"bandwidthLimitationReason"`
			SdkDefaultDeviceManagement   []string `json:"sdkDefaultDeviceManagement"`
			JoinOrderLayout              []string `json:"joinOrderLayout"`
			PinLayout                    []string `json:"pinLayout"`
			SendSelfViewVideoSlot        []string `json:"sendSelfViewVideoSlot"`
			ServerLayoutTransition       []string `json:"serverLayoutTransition"`
			SdkPublisherOptimizeBitrate  []string `json:"sdkPublisherOptimizeBitrate"`
			SdkNetworkLostDetection      []string `json:"sdkNetworkLostDetection"`
			SdkNetworkPathMonitor        []string `json:"sdkNetworkPathMonitor"`
			PublisherVp9                 []string `json:"publisherVp9"`
			SvcMode                      []string `json:"svcMode"`
			SubscriberOfferAsyncAck      []string `json:"subscriberOfferAsyncAck"`
			AndroidBluetoothRoutingFix   []string `json:"androidBluetoothRoutingFix"`
			FixedIceCandidatesPoolSize   []string `json:"fixedIceCandidatesPoolSize"`
			SdkAndroidTelecomIntegration []string `json:"sdkAndroidTelecomIntegration"`
			SetActiveCodecsMode          []string `json:"setActiveCodecsMode"`
			SubscriberDtlsPassiveMode    []string `json:"subscriberDtlsPassiveMode"`
			SvcModes                     []string `json:"svcModes"`
			ReportTelemetryModes         []string `json:"reportTelemetryModes"`
			KeepDefaultDevicesModes      []string `json:"keepDefaultDevicesModes"`
		} `json:"capabilitiesOffer"`
		SdkInfo struct {
			Implementation string `json:"implementation"`
			Version        string `json:"version"`
			UserAgent      string `json:"userAgent"`
			HwConcurrency  int    `json:"hwConcurrency"`
		} `json:"sdkInfo"`
		SdkInitializationID    string `json:"sdkInitializationId"`
		DisablePublisher       bool   `json:"disablePublisher"`
		DisableSubscriber      bool   `json:"disableSubscriber"`
		DisableSubscriberAudio bool   `json:"disableSubscriberAudio"`
	}
	ParticipantMeta struct {
		Name        string `json:"name"`
		Role        string `json:"role"`
		Description string `json:"description"`
		SendAudio   bool   `json:"sendAudio"`
		SendVideo   bool   `json:"sendVideo"`
	}
	ParticipantAttributes struct {
		Name        string `json:"name,omitempty"`
		Role        string `json:"role,omitempty"`
		Description string `json:"description,omitempty"`
	}
)

type (
	PublisherSdpOfferPayload struct {
		PcSeq  int     `json:"pcSeq"`
		Sdp    string  `json:"sdp"`
		Tracks []Track `json:"tracks"`
	}
	Track struct {
		Mid            string   `json:"mid"`
		TransceiverMid string   `json:"transceiverMid"`
		Kind           string   `json:"kind"`
		Priority       int      `json:"priority"`
		Label          string   `json:"label"`
		Codecs         struct{} `json:"codecs"`
		GroupId        int      `json:"groupId"`
		Description    string   `json:"description"`
	}
)

type UpdatePublisherTrackDescriptionPayload struct {
	PublisherTrackDescriptions []Track `json:"publisherTrackDescriptions"`
}

type (
	SdkCodecsInfoPayload struct {
		Vp8  CodecSupport   `json:"vp8"`
		Vp9  []CodecSupport `json:"vp9"`
		Av1  []CodecSupport `json:"av1"`
		H264 []CodecSupport `json:"h264"`
	}
	CodecSupport struct {
		Supported string `json:"supported"`
		HwDecode  string `json:"hwDecode"`
		HwEncode  string `json:"hwEncode"`
		IsoString string `json:"isoString"`
	}
)

type SetSlotsPayload struct {
	Slots []struct {
		Width  int `json:"width"`
		Height int `json:"height"`
	} `json:"slots"`
	AudioSlotsCount    int         `json:"audioSlotsCount"`
	Key                int         `json:"key"`
	ShutdownAllVideo   interface{} `json:"shutdownAllVideo"`
	WithSelfView       bool        `json:"withSelfView"`
	SelfViewVisibility string      `json:"selfViewVisibility"`
	GridConfig         struct {
	} `json:"gridConfig"`
}

// common payloads:

type WebrtcIceCandidatePayload struct {
	Candidate        string `json:"candidate"`
	SdpMid           string `json:"sdpMid"`
	UsernameFragment string `json:"usernameFragment,omitempty"`
	SdpMlineIndex    int    `json:"sdpMlineIndex"`
	Target           string `json:"target"`
	PcSeq            int    `json:"pcSeq"`
}

type AckPayload struct {
	Status struct {
		Code        string `json:"code"`
		Description string `json:"description,omitempty"`
	} `json:"status"`
}

type SdpPayload struct {
	PcSeq int    `json:"pcSeq"`
	Sdp   string `json:"sdp"`
}
