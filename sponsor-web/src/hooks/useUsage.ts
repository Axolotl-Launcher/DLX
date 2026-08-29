import { useEffect, useState } from "react"
import { api } from "@/lib/api"
import type { UsageResponse } from "@/lib/types"
export function useUsage(){const[data,setData]=useState<UsageResponse|null>(null);const[loading,setLoading]=useState(true);useEffect(()=>{let active=true;api.usage().then(value=>{if(active)setData(value)}).catch(()=>{if(active)setData(null)}).finally(()=>{if(active)setLoading(false)});return()=>{active=false}},[]);return{data,loading}}
