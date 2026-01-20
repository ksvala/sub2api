import { apiClient } from '../client'

export interface UploadResponse {
  url: string
}

export async function uploadImage(file: File): Promise<UploadResponse> {
  const formData = new FormData()
  formData.append('file', file)
  const { data } = await apiClient.post<UploadResponse>('/admin/uploads/image', formData, {
    headers: {
      'Content-Type': 'multipart/form-data'
    }
  })
  return data
}

export const uploadsAPI = {
  uploadImage
}

export default uploadsAPI
