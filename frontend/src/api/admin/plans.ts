import { apiClient } from '../client'
import type { PaginatedResponse, Plan } from '@/types'

export interface CreatePlanRequest {
  title: string
  description: string
  price: number
  group_name: string
  group_sort: number
  daily_quota: number
  total_quota: number
  purchase_qr_url: string
  enabled: boolean
  sort_order: number
}

export interface UpdatePlanRequest extends Partial<CreatePlanRequest> {
  id: number
}

export const plansAPI = {
  async getPlans() {
    const { data } = await apiClient.get<PaginatedResponse<Plan>>('/admin/plans', {
      params: { page: 1, page_size: 200 }
    })
    return data.items || []
  },

  async createPlan(plan: CreatePlanRequest) {
    const { data } = await apiClient.post<Plan>('/admin/plans', plan)
    return data
  },

  async updatePlan(id: number, plan: Partial<CreatePlanRequest>) {
    const { data } = await apiClient.put<Plan>(`/admin/plans/${id}`, plan)
    return data
  },

  async deletePlan(id: number) {
    const { data } = await apiClient.delete<{ message: string }>(`/admin/plans/${id}`)
    return data
  }
}

export default plansAPI
