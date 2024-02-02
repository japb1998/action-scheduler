import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet } from '@angular/router';
import { FormsModule } from '@angular/forms';

enum NotificationType {
  Daily = "daily",
  OneTime = "one_time",
  Monthly = "monthly",
}

type CreateSchedule = {
  id: string,
  name: string,
  expression: {
    type: string,
    start_date: string,
    end_date?: string,
  },
  payload: any,
  action: string,
  created_by: string
}

type  Schedule = CreateSchedule & { action: { name: string, id: string } }
@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, RouterOutlet, FormsModule],
  templateUrl: './app.component.html',
  styleUrl: './app.component.css'
})
export class AppComponent implements OnInit {
  title = 'notification-handler';
  payload: any = ""
  type: string = NotificationType.OneTime
  startDate: string = new Date().toISOString().split("T")[0]
  endDate: string = ""
  name: string = "";
  action: string = "1"
  startHour: string = "00:00"
  endHour: string = "23:59"
  endDateBox: boolean = false;

  // no form properties
  schedules: Schedule[] = [];
  payloadError: string = "";
  invalidPayload: boolean = false;


  ngOnInit() {
    this.loadSchedules();
  }
  onPayloadChange(event: any) {
   try {
    this.payload = JSON.stringify(JSON.parse(event.target.value), null, 4);
    this.invalidPayload = false;
   } catch (e) {
    this.invalidPayload = true;
    this.payloadError = (e as Error).message;
    this.payload = event.target.value;
   }
  }

  protected get displayEndDate() {
    switch (this.type) {
      case NotificationType.Monthly:
      case NotificationType.Daily:
        return true
      default:
        return false
    }
  }

  async createSchedule(e: any) {
    e.preventDefault();
    console.log("Creating schedule");

    if (this.startDate === "" || this.name === "" || this.payload === "" || this.invalidPayload) throw new Error("Invalid input");

    const year = this.startDate.split("-")[0];
    const month = this.startDate.split("-")[1];
    const day = this.startDate.split("-")[2];
    const startDate = new Date();

    startDate.setFullYear(parseInt(year));
    startDate.setMonth(parseInt(month) - 1);
    startDate.setDate(parseInt(day));
    startDate.setHours(parseInt(this.startHour.split(":")[0]));
    startDate.setMinutes(parseInt(this.startHour.split(":")[1]));
    startDate.setSeconds(0);
    startDate.setMilliseconds(0);

    const schedule: {
      name: string,
      expression: {
        type: string,
        start_date: string,
        end_date?: string,
      },
      payload: any,
      action: string,
      created_by: string
    } = {
      name: this.name,
      expression: {
        type: this.type,
        start_date: startDate.toISOString(),
      },
      payload: JSON.parse(this.payload),
      action: this.action, 
      created_by: "admin"
    };
    
    if (this.type !== NotificationType.OneTime && this.endDate !== "") {
      const year = this.endDate.split("-")[0];
      const month = this.endDate.split("-")[1];
      const day = this.endDate.split("-")[2];
      const endDate = new Date();
  
      endDate.setFullYear(parseInt(year));
      endDate.setMonth(parseInt(month) - 1);
      endDate.setDate(parseInt(day));
      endDate.setHours(parseInt(this.startHour.split(":")[0]));
      endDate.setMinutes(parseInt(this.startHour.split(":")[1]));
      endDate.setSeconds(0);
      endDate.setMilliseconds(0);
      schedule.expression.end_date = endDate.toISOString();
  
    }
    fetch("http://localhost:8080/schedule", {
      method: "POST",
      body: JSON.stringify(schedule)
    }).then((_) => {
      this.payload = "";
      this.name = "";
      this.startDate = new Date().toISOString().split("T")[0];
      this.endDate = "";
      this.action = "1";
      this.startHour = "00:00";
      this.endHour = "23:59";

      return this.loadSchedules();
    }).then(() => console.log('fetched after creation')).catch(console.error);

  }

  loadSchedules() {
    fetch("http://localhost:8080/schedule?limit=20").then((res) => {
      res.json().then(({ items: schedules }: {
        total: number,
        items: Schedule[],
        page: number,
        limit: number
      }) => {
        this.schedules = schedules;
      })
    }).catch(console.error);
  }

  deleteSchedule(id: string) {
    fetch(`http://localhost:8080/schedule/${id}`, {
      method: "DELETE"
    }).then((res) => {
      this.schedules = this.schedules.filter((schedule) => schedule.id !== id);
    }).catch(console.error);
  }
}
