import { Controller, Get, Post, Delete, HttpCode } from '@nestjs/common';
import { OrderService } from './order.service';

@Controller('orders')
export class OrderController {
  constructor(private readonly orderService: OrderService) {}

  @Get()
  async getOrders() {
    return this.orderService.getAllOrders();
  }

  @Delete()
  @HttpCode(204)
  async deleteAllOrders() {
    await this.orderService.deleteAllOrders();
  }
}
