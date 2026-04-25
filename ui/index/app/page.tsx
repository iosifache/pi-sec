import { headers } from "next/headers"

import { PackagesTable } from "@/components/packages-table"
import { loadPackagesDashboardData } from "@/lib/load-cached-data"

export default async function Page() {
  const requestHeaders = await headers()
  const protocol = requestHeaders.get("x-forwarded-proto") ?? "http"
  const host = requestHeaders.get("x-forwarded-host") ?? requestHeaders.get("host")
  const baseUrl = host ? `${protocol}://${host}` : undefined
  const data = await loadPackagesDashboardData(baseUrl)

  return <PackagesTable data={data} />
}
