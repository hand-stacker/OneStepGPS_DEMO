export interface User {
  user_id: number | string
  email: string
}

export type sort_order = "asc" | "desc"

export interface UserSortPreference {
  user_id: number | string
  sort_order: sort_order
}

export interface HiddenDevice {
    user_id: number | string
    device_id : string
    ignore : boolean
}

export interface DeviceNickname {
    user_id: number | string
    device_id : string
    display_name : string
    ignore : boolean
}

export interface DeviceMarker {
    user_id : number | string
    device_id : string
    content_type : string
    data : string
    ignore : boolean
}