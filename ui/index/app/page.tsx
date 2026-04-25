import { PackagesTable } from "@/components/packages-table"
import { loadPackagesDashboardData } from "@/lib/load-cached-data"

export default async function Page() {
  const data = await loadPackagesDashboardData()

  return <PackagesTable data={data} />
}
