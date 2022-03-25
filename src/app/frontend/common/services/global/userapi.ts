import { Injectable } from '@angular/core'
import {HttpClient} from '@angular/common/http';

export interface DashboardUser {
  id:number;
  username: string;
  password: string;
  token: string;
  type: string;
}

@Injectable({
  providedIn: 'root'
})

export class UserApi {
  constructor(
    private http:HttpClient,
  ){}

  allUsers()
  {
    return this.http.get<DashboardUser[]>("/api/v1/users")
  }
}
