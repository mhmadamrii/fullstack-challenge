import { Injectable } from "@nestjs/common";
import { CreateProductDto } from "./product.dto";
import { PrismaService } from "src/prisma.service";

@Injectable()
export class ProductService {
  constructor(private prisma: PrismaService){}

  async createProduct(dto: CreateProductDto) {
    return this.prisma.product.create({data: dto})
  }

  async getAllProducts() {
    return this.prisma.product.findMany()
  }

  async getProductById(id: string) {
    return this.prisma.product.findUnique({ where: { id } });
  }
}