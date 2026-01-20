/**
 * Admin Invite API endpoints
 */

import { apiClient } from '../client'
import type { InviteLog, InviteRecord, InviteSettings, PaginatedResponse } from '@/types'

export interface InviteLogQuery {
  page?: number
  page_size?: number
  action?: string
  inviter_id?: number
  invitee_id?: number
  inviter_email?: string
  invitee_email?: string
  start_time?: string
  end_time?: string
}

export async function getInviteSettings(): Promise<InviteSettings> {
  const { data } = await apiClient.get<InviteSettings>('/admin/invites/settings')
  return data
}

export async function updateInviteSettings(settings: InviteSettings): Promise<InviteSettings> {
  const { data } = await apiClient.put<InviteSettings>('/admin/invites/settings', settings)
  return data
}

export async function listInviteLogs(query: InviteLogQuery): Promise<PaginatedResponse<InviteLog>> {
  const { data } = await apiClient.get<PaginatedResponse<InviteLog>>('/admin/invites/logs', {
    params: query
  })
  return data
}

export async function confirmInvite(inviteeId: number): Promise<InviteRecord> {
  const { data } = await apiClient.post<InviteRecord>(`/admin/invites/${inviteeId}/confirm`)
  return data
}

export const invitesAdminAPI = {
  getInviteSettings,
  updateInviteSettings,
  listInviteLogs,
  confirmInvite
}

export default invitesAdminAPI
