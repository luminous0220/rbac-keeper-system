package dept

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"rbackeeper/global"
	"rbackeeper/model"
	"rbackeeper/response"
	"rbackeeper/utils"
)

// Tree 部门树
func Tree(ctx *gin.Context, query DeptTreeQuery) ([]model.Dept, error) {
	depts, err := GetAllDeptsScoped(ctx, query.Status)
	if err != nil {
		return nil, err
	}
	if query.DeptName != "" {
		depts = utils.FilterTreeWithAncestors(depts,
			func(d model.Dept) uint { return d.ID },
			func(d model.Dept) uint { return d.ParentID },
			func(d model.Dept) bool { return strings.Contains(d.DeptName, query.DeptName) },
		)
	}
	tree := utils.Build(depts,
		func(d model.Dept) uint { return d.ID },
		func(d model.Dept) uint { return d.ParentID },
		func(p *model.Dept, kids []*model.Dept) { p.Children = kids },
	)
	return tree, nil
}

// Add 新增部门（父部门须在当前用户数据权限内）
func Add(ctx *gin.Context, body DeptAddBody) error {
	if body.ParentID != 0 {
		if ok, err := ExistsByIDScoped(ctx, global.DB, &body.ParentID); err != nil {
			return err
		} else if !ok {
			return response.NewBizError(response.ErrInvalidParam, "父部门不存在")
		}
	}

	if exist, err := ExistsByDeptName(global.DB, body.DeptName, 0); err != nil {
		return err
	} else if exist {
		return response.NewBizError(response.ErrInvalidParam, "部门名称重复")
	}

	d := &model.Dept{
		ParentID: body.ParentID,
		DeptName: body.DeptName,
		OrderNum: body.OrderNum,
		Leader:   body.Leader,
		Phone:    body.Phone,
		Email:    body.Email,
		Status:   1, // 默认正常
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		return Create(tx, d)
	})
}

// Edit 修改部门
func Edit(ctx *gin.Context, id uint, body DeptEditBody) error {
	// 检查id是否正确（目标部门须在数据权限内，不在权限内同样返回"部门不存在"，不泄露行存在性）
	d, err := GetByIDScoped(ctx, global.DB, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewBizError(response.ErrNotFound, "部门不存在")
		}
		return err
	}

	// 唯一性检查
	if body.DeptName != nil {
		if exist, err := ExistsByDeptName(global.DB, *body.DeptName, id); err != nil {
			return err
		} else if exist {
			return response.NewBizError(response.ErrInvalidParam, "部门名称重复")
		}
	}

	updates := map[string]any{}
	if body.DeptName != nil {
		updates["dept_name"] = *body.DeptName
	}
	if body.Leader != nil {
		updates["leader"] = *body.Leader
	}
	if body.OrderNum != nil {
		updates["order_num"] = *body.OrderNum
	}
	if body.Phone != nil {
		updates["phone"] = *body.Phone
	}
	if body.Email != nil {
		updates["email"] = *body.Email
	}
	if body.Status != nil {
		updates["status"] = *body.Status
	}
	if body.Remark != nil {
		updates["remark"] = *body.Remark
	}

	// 父部门变更：
	if body.ParentID != nil {
		newParent := *body.ParentID
		if newParent != 0 {
			// 检查父部门是否存在，且在当前用户数据权限内
			if ok, err := ExistsByIDScoped(ctx, global.DB, &newParent); err != nil {
				return err
			} else if !ok {
				return response.NewBizError(response.ErrInvalidParam, "父部门不存在")
			}
			// 环检测必须用全量部门（不应用数据权限），否则环可能藏在权限范围外的节点里
			all, err := GetAllDeptsScoped(ctx, nil)
			if err != nil {
				return err
			}
			// 环检测：自身不能成为自身子部门的子节点
			parentMap := make(map[uint]uint, len(all)) // 创建 dept id->parentID 的映射表
			for i := range all {
				parentMap[all[i].ID] = all[i].ParentID
			}
			if utils.WouldCreateCycle(parentMap, d.ID, newParent) {
				return response.NewBizError(response.ErrInvalidParam, "不能将部门挂载到自身或其子部门下")
			}
		}
		updates["parent_id"] = newParent
	}

	return global.DB.Transaction(func(tx *gorm.DB) error {
		return UpdateByID(tx, d.ID, updates)
	})
}

// Delete 删除部门（目标部门须在当前用户数据权限内，不在权限内同样返回"部门不存在"）
func Delete(ctx *gin.Context, id uint) error {
	if ok, err := ExistsByIDScoped(ctx, global.DB, &id); err != nil {
		return err
	} else if !ok {
		return response.NewBizError(response.ErrNotFound, "部门不存在")
	}
	cnt, err := CountByParentID(global.DB, id)
	if err != nil {
		return err
	}
	if cnt > 0 {
		return response.NewBizError(response.ErrForbidden, "存在子部门，无法删除")
	}
	userCnt, err := CountUsersByDeptID(global.DB, id)
	if err != nil {
		return err
	}
	if userCnt > 0 {
		return response.NewBizError(response.ErrForbidden, "部门下存在用户，无法删除")
	}
	return global.DB.Transaction(func(tx *gorm.DB) error {
		return DeleteByID(tx, id)
	})
}
