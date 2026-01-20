/**
 * Invite API endpoints (user)
 */

import { apiClient } from './client'
import type { InviteSummary, InviteRecord, InviteRewardRecord, PaginatedResponse } from '@/types'

export interface InviteRecordsQuery {
  page?: number
  page_size?: number
  status?: string
}

export interface InviteRewardsQuery {
  page?: number
  page_size?: number
}

export async function getInviteSummary(): Promise<InviteSummary> {
  const { data } = await apiClient.get<InviteSummary>('/invites/summary')
  return data
}

export async function listInviteRecords(query: InviteRecordsQuery): Promise<PaginatedResponse<InviteRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<InviteRecord>>('/invites/records', {
    params: query
  })
  return data
}

export async function listInviteRewards(query: InviteRewardsQuery): Promise<PaginatedResponse<InviteRewardRecord>> {
  const { data } = await apiClient.get<PaginatedResponse<InviteRewardRecord>>('/invites/rewards', {
    params: query
  })
  return data
}

export const inviteAPI = {
  getInviteSummary,
  listInviteRecords,
  listInviteRewards
}

export default inviteAPI
